package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MGMCN/P2PFileSharing/p2p"
)

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Peer-to-peer file sharing over LAN.\n")
		fmt.Printf("Usage: Plz see \n")
		os.Exit(0)
	}

	var err error
	runtimeErrChan := make(chan error, 10)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	p2pNode := p2p.Newp2pNode()
	err = p2pNode.InitP2PNode(ctx, cfg.RendezvousString, cfg.listenHost, cfg.listenPort, cfg.sharedDirectory, runtimeErrChan)
	if err != nil {
		os.Exit(1)
	} else {
		log.Printf("Peer listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)
	}

	for {
		select {
		case runtimeErr := <-runtimeErrChan:
			log.Printf("Runtime error occurs! %s\n", runtimeErr)
			os.Exit(1)
		case <-sigCh:
			// Do something before disconnect
			// Finally
			cancel()
			log.Printf("Peer ends gracefully!\n")
			os.Exit(0)
		}
	}
}
