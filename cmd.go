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
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	from = kingpin.Flag("from", "from address:port").Short('f').String()
	to   = kingpin.Flag("to", "to address:port").Short('t').String()
	list = kingpin.Flag("list", "list local addresses").Short('l').Bool()

	city     = kingpin.Flag("city", "only accept incoming connections that originate from given city").String()
	region   = kingpin.Flag("region", "only accept incoming connections that originate from given region (eg: state)").String()
	country  = kingpin.Flag("country", "only accept incoming connections that originate from given 2 letter country abbreviation").String()
	distance = kingpin.Flag("distance", "only accept incoming connections from within the distance (in miles)").Short('d').Int()
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

func tcpStart(from string, to string) {
	proto := "tcp"

	fromAddress, err := net.ResolveTCPAddr(proto, from)
	errHandler(err, true)

	//validateLocation(fromAddress,nil)

	toAddress, err := net.ResolveTCPAddr(proto, to)
	errHandler(err, true)

	listener, err := net.ListenTCP(proto, fromAddress)
	errHandler(err, true)

	defer listener.Close()

	logger.Infof("Forwarding %s traffic from '%v' to '%v'", proto, fromAddress, toAddress)

	for {
		src, err := listener.Accept()
		errHandler(err, true)
		logger.Infof("New connection established from '%v'", src.RemoteAddr())
		go fwd(src, to, proto)
	}
}

func main() {
	kingpin.Parse()
	if *list {
		nics()
		os.Exit(0)
	}

	if !(len(*from) >= 9 && len(*to) >= 9) {
		kingpin.Usage()
		os.Exit(1)
	}

	loggingHandler()
	signalHandler()
	tcpStart(*from, *to)
}
