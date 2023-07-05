package cli

import (
	"github.com/MGMCN/P2PFileSharing/pkg/handler"
	"log"
)

type PeerCli struct{}

func NewPeerCli() *PeerCli {
	return &PeerCli{}
}

func (pc *PeerCli) InitPeerCli() {}

func (pc *PeerCli) Execute(commands []string, handler handler.BaseStreamHandler) {
	if handler != nil {
		go func() {
			errs := handler.OpenStreamAndSendRequest(commands)
			if len(errs) != 0 {
				log.Printf("Some errors occurred while executing %s\n", commands)
			}
		}()
	} else {
		log.Println("Please make sure you are entering a valid peer command")
	}
}
