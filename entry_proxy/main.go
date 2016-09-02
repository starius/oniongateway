package main

/*  Entry Proxy for Onion Gateway

See also:

  * https://github.com/DonnchaC/oniongateway/blob/master/docs/design.rst#32-entry-proxy
  * https://gist.github.com/Yawning/bac58e08a05fc378a8cc (SOCKS5 client, Tor)
  * https://habrahabr.ru/post/142527/ (TCP proxy)
*/

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		proxyNet = flag.String(
			"proxyNet",
			"tcp",
			"Proxy network type",
		)
		proxyAddr = flag.String(
			"proxyAddr",
			"127.0.0.1:9050",
			"Proxy address",
		)
		entryProxy = flag.String(
			"entry-proxy",
			":443",
			"host:port of entry proxy",
		)
		httpRedirect = flag.String(
			"http-redirect",
			":80",
			"host:port of redirecting HTTP server ('' to disable)",
		)
		onionPort = flag.Int(
			"onion-port",
			443,
			"Port on onion site to use",
		)
		cacheMb = flag.Int(
			"cache-mb",
			10,
			"In-memory cache size for host->onion map (MB)",
		)
	)

	flag.Parse()

	if *httpRedirect != "" {
		redirectingServer, err := NewRedirect(*httpRedirect, *entryProxy)
		if err != nil {
			fmt.Printf("Unable to create redirecting HTTP server: %s\n", err)
			os.Exit(1)
		}
		go redirectingServer.ListenAndServe()
	}

	cacheBytes := *cacheMb * 1024 * 1024
	proxy := NewTLSProxy(*onionPort, *proxyNet, *proxyAddr, cacheBytes)
	proxy.Listen("tcp", *entryProxy)
	proxy.Start()
}
