package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/p2p"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	help, cfg := parseFlags()

	if *help {
		fmt.Printf("Peer-to-peer file sharing over LAN.\n")
		flag.PrintDefaults()
	} else {
		var hostID string
		var err error
		runtimeErrChan := make(chan error, 10)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		ctx := context.Background()

		p2pNode := p2p.Newp2pNode()
		err, hostID = p2pNode.InitP2PNode(ctx, cfg.RendezvousString, cfg.listenHost, cfg.listenPort, cfg.sharedDirectory, runtimeErrChan)
		if err != nil {
			log.Println("InitP2PNode error: node ends gracefully")
		} else {
			log.Printf("Peer listening on: %s with port: %d hostID: %s\n", cfg.listenHost, cfg.listenPort, hostID)

			for {
				select {
				case runtimeErr := <-runtimeErrChan:
					// We should handle this error
					log.Printf("Runtime error occurs! %s\n", runtimeErr)
				case <-sigCh:
					p2pNode.Leave()
				}
			}
		}
	}
}
