package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	from = kingpin.Flag("from", "from address:port").Short('f').String()
	to   = kingpin.Flag("to", "to address:port").Short('t').String()
	list = kingpin.Flag("list", "list local addresses").Short('l').Bool()
)

func errHandler(err error, fatal bool) {
	if err != nil {
		color.Set(color.FgRed)
		fmt.Fprintf(os.Stderr, "[Error] %s\n", err.Error())
		color.Unset()
		if fatal {
			os.Exit(1)
		}
	}
}

func fwd(src net.Conn, remote string, proto string) {
	dst, err := net.Dial(proto, remote)
	errHandler(err, false)
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

	localAddress, err := net.ResolveTCPAddr(proto, from)
	errHandler(err, true)

	remoteAddress, err := net.ResolveTCPAddr(proto, to)
	errHandler(err, true)

	listener, err := net.ListenTCP(proto, localAddress)
	errHandler(err, true)

	defer listener.Close()

	fmt.Printf("Forwarding %s traffic from '%v' to '%v'\n", proto, localAddress, remoteAddress)
	color.Set(color.FgYellow)
	fmt.Println("<CTRL+C> to exit")
	fmt.Println()
	color.Unset()

	for {
		src, err := listener.Accept()
		errHandler(err, true)
		fmt.Printf("New connection established from '%v'\n", src.RemoteAddr())
		go fwd(src, to, proto)
	}
}

func ctrlc() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		color.Set(color.FgGreen)
		fmt.Println("\nExecution stopped by", sig)
		color.Unset()
		os.Exit(0)
	}()
}

func main() {
	kingpin.Parse()
	if *list {
		nics()
		os.Exit(0)
	}
	fmt.Printf("from, to: %s %s\n", *from, *to)

	ctrlc()
	tcpStart(*from, *to)
}
