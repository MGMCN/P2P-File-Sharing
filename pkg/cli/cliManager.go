package cli

import (
	"github.com/MGMCN/P2PFileSharing/pkg/handler"
	"log"
)

type CommandLineInterfaceManager struct {
	cli map[string]BaseCli
}

func NewCliManager() *CommandLineInterfaceManager {
	return &CommandLineInterfaceManager{}
}

func (clim *CommandLineInterfaceManager) InitCliManager() {
	clim.cli = make(map[string]BaseCli)

	peerCli := NewPeerCli()
	peerCli.InitPeerCli()
	clim.cli["peer"] = peerCli

	cacheCli := NewCacheCli()
	cacheCli.InitCacheCli()
	clim.cli["cache"] = cacheCli
}

func (clim *CommandLineInterfaceManager) Execute(commands []string, handler handler.BaseStreamHandler) {
	if cli, ok := clim.cli[commands[0]]; ok {
		cli.Execute(commands, handler)
	} else {
		log.Println("No corresponding cli handler found")
	}
}
