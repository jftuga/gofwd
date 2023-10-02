/*
cmd.go
-John Taylor
April 11 2020

A cross-platform TCP port forwarder with Duo and GeoIP integration

The following functions were adopted from: https://github.com/kintoandar/fwd/
errHandler, signalHandler, fwd, tcpStart

*/

package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "0.7.2"

var (
	list        = kingpin.Flag("int", "list local interface IP addresses").Short('i').Bool()
	from        = kingpin.Flag("from", "from IP address:port; use 'MAIN' for the address portion to use system's primary interface").Short('f').String()
	to          = kingpin.Flag("to", "to IP address:port").Short('t').String()
	examples    = kingpin.Flag("examples", "show command line example and then exit").Bool()
	versionOnly = kingpin.Flag("version", "show version and then exit").Bool()

	city      = kingpin.Flag("city", "only accept incoming connections that originate from given city").String()
	region    = kingpin.Flag("region", "only accept incoming connections that originate from given region (eg: state)").String()
	country   = kingpin.Flag("country", "only accept incoming connections that originate from given 2 letter country abbreviation").String()
	loc       = kingpin.Flag("loc", "only accept from within a geographic radius; format: LATITUDE,LONGITUDE (use with --distance)").Short('l').String()
	distance  = kingpin.Flag("distance", "only accept from within a given distance (in miles)").Short('d').Float64()
	allowCIDR = kingpin.Flag("allow", "allow from a comma delimited list of CIDR networks, bypassing geo-ip, duo").Short('A').String()
	denyCIDR  = kingpin.Flag("deny", "deny from a comma delimited list of CIDR networks, disregarding geo-ip, duo").Short('D').String()

	duo              = kingpin.Flag("duo", "path to duo ini config file and duo username; format: filename:user (see --examples)").String()
	duoAuthCacheTime = kingpin.Flag("duo-cache-time", "number of seconds to cache a successful Duo authentication (default is 120)").Default("120").Int64()
	private          = kingpin.Flag("private", "allow RFC1918 private addresses for the remote IP").Short('p').Bool()
)

var logger *zap.SugaredLogger

func errHandler(err error, fatal bool) {
	if err != nil {
		logger.Warnf(err.Error())
		if fatal {
			os.Exit(1)
		}
	}
}

func signalHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Errorf("Execution stopped by %s", sig)
		os.Exit(0)
	}()
}

func loggingHandler() {
	cfg := zap.Config{
		Encoding:    "console",
		OutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	loggerPlain, _ := cfg.Build()
	logger = loggerPlain.Sugar()
}

// IsPrivateIPv4 https://gist.github.com/r4um/5986319
func isPrivateIPv4(s string) bool {
	var rfc1918 []string = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	var ip net.IP = net.ParseIP(s)

	if ip == nil {
		return false
	}

	for _, cidr := range rfc1918 {
		_, net, _ := net.ParseCIDR(cidr)
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

func fwd(src net.Conn, remote string, proto string) {
	dst, err := net.Dial(proto, remote)
	errHandler(err, false)
	if err != nil {
		return
	}
	go func() {
		_, err = io.Copy(src, dst)
		errHandler(err, false)
	}()
	go func() {
		_, err = io.Copy(dst, src)
		errHandler(err, false)
	}()
}

func validateCIDRList(all *string) (string, bool) {
	networks := strings.Split(*all, ",")
	for _, cidr := range networks {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return cidr, false
		}
	}
	return "", true
}

func ipIsInCIDR(s string, cidr *string) bool {
	ip := net.ParseIP(s)
	networks := strings.Split(*cidr, ",")
	for _, cidr := range networks {
		_, ipv4Net, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Warnf("Invalid CIDR network: %s", err)
			return false
		}
		if ipv4Net.Contains(ip) {
			return true
		}
	}
	return false
}

func tcpStart(from string, to string, localGeoIP ipInfoResult, restrictionsGeoIP ipInfoResult, duoCred duoCredentials, duoAuthCacheTime int64, allowPrivateIP bool) {
	proto := "tcp"

	fromAddress, err := net.ResolveTCPAddr(proto, from)
	errHandler(err, true)

	toAddress, err := net.ResolveTCPAddr(proto, to)
	errHandler(err, true)

	listener, err := net.ListenTCP(proto, fromAddress)
	errHandler(err, true)
	defer listener.Close()

	logger.Infof("[%v] Forwarding to [%s] [%s]", fromAddress, proto, toAddress)
	for {
		src, err := listener.Accept()
		errHandler(err, true)

		slots := strings.Split(src.RemoteAddr().String(), ":")
		remoteIP := slots[0]
		logger.Infof("[%v] Incoming connection initiated", remoteIP)
		remoteGeoIP, err := getIPInfo(remoteIP)
		if "127.0.0.1" != remoteIP {
			if allowPrivateIP && isPrivateIPv4(remoteIP) {
				logger.Infof("[%v] allowing private IPv4 address, skip lat,lon checks", remoteIP)
				err = nil
			}
			if err != nil {
				logger.Warnf("%s", err)
				continue
			}
		}

		if len(*denyCIDR) > 0 && ipIsInCIDR(remoteIP, denyCIDR) {
			logger.Infof("[%v] DENIED; Explicitly Denied by -D option", src.RemoteAddr())
			continue
		}

		if len(*allowCIDR) > 0 && ipIsInCIDR(remoteIP, allowCIDR) {
			logger.Infof("[%v] ESTABLISHED; Explicitly Allowed by -A option", src.RemoteAddr())
			go fwd(src, to, proto)
			continue
		}
		invalidLocation, distanceCalc := validateLocation(localGeoIP, remoteGeoIP, restrictionsGeoIP)
		if "127.0.0.1" != remoteIP {
			if allowPrivateIP && isPrivateIPv4(remoteIP) {
				logger.Infof("[%v] allowing private IPv4 address, skip loc,dist checks", remoteIP)
				invalidLocation = ""
			}
			if len(invalidLocation) > 0 {
				logger.Warnf("%s %s", invalidLocation, distanceCalc)
				// do not attempt: listener.Close()
				continue
			}
		}

		if len(duoCred.name) > 0 {
			var allowed bool
			lastAuthTime := "(never)"
			cachedDuoAuth := ""
			if duoCred.lastAuthTime > 0 {
				lastAuthTime = fmt.Sprintf("%v", time.Unix(duoCred.lastAuthTime, 0))
			}
			logger.Infof("[%s] last auth time: %v", duoCred.name, lastAuthTime)

			current := time.Now().Unix()
			diff := current - duoCred.lastAuthTime
			if diff <= duoAuthCacheTime && duoCred.lastIP == remoteIP {
				logger.Infof("[%s] last auth time was only %v seconds ago, will not ask again", duoCred.name, diff)
				cachedDuoAuth = " CACHED"
			} else {
				allowed, err = duoCheck(duoCred)
				if err != nil {
					errHandler(err, false)
					logger.Warnf("[%v] DENIED; Duo Auth for user: %s", src.RemoteAddr(), duoCred.name)
					continue
				}
				if !allowed {
					errHandler(errors.New("Duo Auth returned false"), false)
					logger.Warnf("[%v] DENIED; Duo Auth for user: %s", src.RemoteAddr(), duoCred.name)
					continue
				}
				duoCred.lastAuthTime = time.Now().Unix()
				duoCred.lastIP = remoteIP
			}
			logger.Infof("[%v] ACCEPTED%s; Duo Auth for user: %s", src.RemoteAddr(), cachedDuoAuth, duoCred.name)
		}

		logger.Infof("[%v] ESTABLISHED; %s", src.RemoteAddr(), distanceCalc)
		go fwd(src, to, proto)
	}
}

func showExamples() {
	examples := getExamples()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Example", "Command"})

	for _, entry := range examples {
		table.Append(entry)
	}
	table.Render()
}

func getDuoConfig(duoFile string, duoUser string, duoAuthCacheTime int64) duoCredentials {
	duoCred, err := duoReadConfig(duoFile, duoUser)
	if err != nil {
		errHandler(err, true)
		os.Exit(1)
	}
	logger.Infof("Duo auth activated for user: %s; cache time: %v seconds", duoCred.name, duoAuthCacheTime)
	return duoCred
}

func main() {
	loggingHandler()
	signalHandler()

	kingpin.Parse()

	if *versionOnly {
		fmt.Fprintf(os.Stderr, "gofwd, version %s\n", version)
		fmt.Fprintf(os.Stderr, "https://github.com/jftuga/gofwd\n\n")
		os.Exit(0)
	}

	if *list {
		nics()
		os.Exit(0)
	}

	if *examples {
		showExamples()
		os.Exit(0)
	}

	if !(len(*from) >= 9 && len(*to) >= 9) {
		kingpin.FatalUsage("Both --from and --to are mandatory")
		os.Exit(1)
	}

	if *from == *to {
		kingpin.FatalUsage("--from and --to can not be identical")
		os.Exit(1)
	}

	if strings.HasPrefix(*from, "MAIN:") {
		*from = strings.Replace(*from, "MAIN", getMainNic(), 1)
	}

	if len(*loc) > 0 && 0 == *distance {
		kingpin.FatalUsage("--distance must be used with --loc")
		os.Exit(1)
	}

	if *distance > 0 && (len(*city) > 0 || len(*region) > 0 || len(*country) > 0) {
		kingpin.FatalUsage("--distance can not be used with any of these: city, region, country; Instead, use --loc with --distance")
		os.Exit(1)
	}

	var badCIDR string
	var ok bool
	if len(*denyCIDR) > 0 {
		badCIDR, ok = validateCIDRList(denyCIDR)
		if !ok {
			kingpin.FatalUsage("Invalid CIDR given for -D option: %s\n", badCIDR)
			os.Exit(1)
		}
	}

	if len(*allowCIDR) > 0 {
		badCIDR, ok = validateCIDRList(allowCIDR)
		if !ok {
			kingpin.FatalUsage("Invalid CIDR given for -A option: %s\n", badCIDR)
			os.Exit(1)
		}
	}

	logger.Infof("gofwd, version %v started", version)

	var duoCred duoCredentials
	if len(*duo) > 0 {
		slots := strings.Split(*duo, ":")
		if len(slots) != 2 {
			kingpin.FatalUsage("Invalid duo filename / user combination")
			os.Exit(1)
		}
		duoCred = getDuoConfig(slots[0], slots[1], *duoAuthCacheTime)
	}

	var restrictionsGeoIP ipInfoResult
	restrictionsGeoIP.City = *city
	restrictionsGeoIP.Region = *region
	restrictionsGeoIP.Country = *country
	restrictionsGeoIP.Distance = *distance
	restrictionsGeoIP.Loc = *loc
	logger.Infof("Geo IP Restrictions: %v", restrictionsGeoIP)

	var localGeoIP ipInfoResult
	var err error
	localGeoIP, err = getIPInfo("")
	if err != nil {
		errHandler(err, true)
		os.Exit(1)
	}

	tcpStart(*from, *to, localGeoIP, restrictionsGeoIP, duoCred, *duoAuthCacheTime, *private)
}
