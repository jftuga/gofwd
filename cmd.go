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

	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.3.0"

// number of seconds to cache a successful Duo authentication
const duoAuthCacheTime int64 = 120

var (
	list        = kingpin.Flag("int", "list local interface IP addresses").Short('i').Bool()
	from        = kingpin.Flag("from", "from address:port").Short('f').String()
	to          = kingpin.Flag("to", "to address:port").Short('t').String()
	examples    = kingpin.Flag("examples", "show command line example and then exit").Bool()
	versionOnly = kingpin.Flag("version", "show version and then exit").Bool()

	city     = kingpin.Flag("city", "only accept incoming connections that originate from given city").String()
	region   = kingpin.Flag("region", "only accept incoming connections that originate from given region (eg: state)").String()
	country  = kingpin.Flag("country", "only accept incoming connections that originate from given 2 letter country abbreviation").String()
	loc      = kingpin.Flag("loc", "only accept incoming connections from within a geographic radius given in LAT,LON").Short('l').String()
	distance = kingpin.Flag("distance", "only accept incoming connections from within the distance (in miles)").Short('d').Float64()

	duo = kingpin.Flag("duo", "path to duo ini config file and duo username; format: filename:user (see --examples)").String()
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

func tcpStart(from string, to string, localGeoIP ipInfoResult, restrictionsGeoIP ipInfoResult, duoCred duoCredentials) {
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
		remoteGeoIP, err := getIPInfo(remoteIP)
		if err != nil {
			logger.Warnf("%s", err)
			continue
		}

		invalidLocation, distanceCalc := validateLocation(localGeoIP, remoteGeoIP, restrictionsGeoIP)
		if len(invalidLocation) > 0 {
			logger.Warnf("%s %s", invalidLocation, distanceCalc)
			// do not attempt: listener.Close()
			continue
		}

		if len(duoCred.name) > 0 {
			var allowed bool
			logger.Infof("[%s] last auth time: %v", duoCred.name, duoCred.lastAuthTime)
			current := time.Now().Unix()
			diff := current - duoCred.lastAuthTime
			if diff <= duoAuthCacheTime {
				logger.Infof("[%s] last auth time was only %v seconds ago, will not ask again", duoCred.name, diff)
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
			}
			logger.Infof("[%v] ACCEPTED; Duo Auth for user: %s", src.RemoteAddr(), duoCred.name)
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

func getDuoConfig(duoFile string, duoUser string) duoCredentials {
	duoCred, err := duoReadConfig(duoFile, duoUser)
	if err != nil {
		errHandler(err, true)
		os.Exit(1)
	}
	logger.Infof("Duo auth activated for user: %s", duoCred.name)
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
	if len(*loc) > 0 && 0 == *distance {
		kingpin.FatalUsage("--distance must also be used with -loc")
		os.Exit(1)
	}

	if *distance > 0 && (len(*city) > 0 || len(*region) > 0 || len(*country) > 0) {
		kingpin.FatalUsage("--distance and not be used with any of these: city, region, country; Instead, use --loc with --distance")
		os.Exit(1)
	}

	var duoCred duoCredentials
	if len(*duo) > 0 {
		slots := strings.Split(*duo, ":")
		if len(slots) != 2 {
			kingpin.FatalUsage("Invalid duo filename / user combination")
			os.Exit(1)
		}
		duoCred = getDuoConfig(slots[0], slots[1])
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

	tcpStart(*from, *to, localGeoIP, restrictionsGeoIP, duoCred)
}
