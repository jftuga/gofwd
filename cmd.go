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
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.2.0"

var (
	list     = kingpin.Flag("int", "list local interface IP addresses").Short('i').Bool()
	from     = kingpin.Flag("from", "from address:port").Short('f').String()
	to       = kingpin.Flag("to", "to address:port").Short('t').String()
	examples = kingpin.Flag("examples", "show command line example").Bool()

	city     = kingpin.Flag("city", "only accept incoming connections that originate from given city").String()
	region   = kingpin.Flag("region", "only accept incoming connections that originate from given region (eg: state)").String()
	country  = kingpin.Flag("country", "only accept incoming connections that originate from given 2 letter country abbreviation").String()
	loc      = kingpin.Flag("loc", "only accept incoming connections from within a geographic radius given in LAT,LON").Short('l').String()
	distance = kingpin.Flag("distance", "only accept incoming connections from within the distance (in miles)").Short('d').Float64()

	duo = kingpin.Flag("duo", "path to duo ini config file").String()
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

func tcpStart(from string, to string, localGeoIP ipInfoResult, restrictionsGeoIP ipInfoResult) {
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
		//logger.Infof("[%v] New connection established", src.RemoteAddr())

		slots := strings.Split(src.RemoteAddr().String(), ":")
		remoteIP := slots[0]
		remoteGeoIP, _ := getIPInfo(remoteIP) /* FIXME: check the err */
		invalidLocation, distanceCalc := validateLocation(localGeoIP, remoteGeoIP, restrictionsGeoIP)
		if len(invalidLocation) > 0 {
			//errHandler(errors.New(invalidLocation), false)
			logger.Warnf("%s %s", invalidLocation, distanceCalc)
			// do not attempt: listener.Close()
			continue
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

func main() {
	kingpin.Parse()
	if *list {
		nics()
		os.Exit(0)
	}

	if *examples {
		showExamples()
		os.Exit(0)
	}

	if len(*duo) > 0 {
		duoReadConfig(*duo)
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

	//FIXME: do not allow -d with any of these: city, region, country

	loggingHandler()
	signalHandler()

	var localGeoIP ipInfoResult
	localGeoIP, _ = getIPInfo("")

	var restrictionsGeoIP ipInfoResult

	restrictionsGeoIP.City = *city
	restrictionsGeoIP.Region = *region
	restrictionsGeoIP.Country = *country
	restrictionsGeoIP.Distance = *distance
	restrictionsGeoIP.Loc = *loc

	logger.Infof("Geo IP Restrictions: %v", restrictionsGeoIP)

	tcpStart(*from, *to, localGeoIP, restrictionsGeoIP)
}
