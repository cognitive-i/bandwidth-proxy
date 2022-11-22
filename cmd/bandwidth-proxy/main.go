package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/signal"

	"github.com/cognitive-i/bandwidth-proxy/proxy"
	"golang.org/x/sys/unix"
)

func main() {
	daemon := flag.Bool("service", false, "run proxy service")
	proxyAddress := flag.String("proxy-address", ":8080", "address:port to listen on")
	controlAddress := flag.String("control-address", ":8081", "address:port to listen on for HTML interface")
	maxBitrate := flag.Int("max-bitrate", 0, "constrain bps, 0 = unlimited")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	defer cancel()

	if *daemon {
		proxyListener := Must(net.Listen("tcp", *proxyAddress))
		controlListener := Must(net.Listen("tcp", *controlAddress))

		proxy.RunProxy(ctx, proxyListener, controlListener, proxy.SetMaxBitrate(*maxBitrate))

	} else {
		proxy.Client(*controlAddress).SetMaxBitrate(*maxBitrate)
	}
}

func Must[T any](result T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return result
}
