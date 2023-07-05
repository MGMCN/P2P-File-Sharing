package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/cli"
	"github.com/MGMCN/P2PFileSharing/pkg/handler"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"strings"
)

type p2pNode struct {
	ctx                context.Context
	peerHost           host.Host
	RendezvousString   string
	listenHost         string
	listenPort         int
	nodeDiscoveryChan  chan peer.AddrInfo
	handlerManager     *handler.Manager
	stdReader          *bufio.Reader
	commandChan        chan string
	runtimeErrChan     chan error
	ourSharedDirectory string
	cache              *runtime.Cache
	cli                *cli.CommandLineInterfaceManager
}

func Newp2pNode() *p2pNode {
	return &p2pNode{}
}

func (p *p2pNode) InitP2PNode(ctx context.Context, RendezvousString string, listenHost string, listenPort int, ourSharedDirectory string, runtimeErrChan chan error) (error, string) {
	var err error
	var prvKey crypto.PrivKey
	var sourceMultiAddr multiaddr.Multiaddr
	p.ctx = ctx
	p.RendezvousString = RendezvousString
	p.listenHost = listenHost
	p.listenPort = listenPort
	p.ourSharedDirectory = ourSharedDirectory
	p.runtimeErrChan = runtimeErrChan
	p.commandChan = make(chan string, 10)
	// Creates a new RSA key pair for this host.
	r := rand.Reader
	prvKey, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		log.Printf("Generate prvKey error! %s\n", err)
	} else {
		// 0.0.0.0 will listen on any interface device.
		sourceMultiAddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", p.listenHost, p.listenPort))
		if err != nil {
			log.Printf("Generate multiaddr error! %s\n", err)
		} else {
			// libp2p.New constructs a new libp2p Host.
			// Other options can be added here.
			p.peerHost, err = libp2p.New(
				libp2p.ListenAddrs(sourceMultiAddr),
				libp2p.Identity(prvKey),
			)
			if err != nil {
				log.Printf("Constructs a new libp2p Host error! %s\n", err)
			} else {
				err = p.initCacheStorage()
				if err != nil {
					log.Printf("initCacheStorage error!\n")
				} else {
					p.initCli()
					p.bindReceiverHandler()
					go p.pollingNodeJoinListener()
					go p.pollingStdinCommandListener()
					go p.startCommandExecutor() // Sender handler will be triggered here by our commands through handlerManager
				}
			}
		}
	}
	return err, p.peerHost.ID().String()
}

func (p *p2pNode) initCacheStorage() error {
	p.cache = runtime.GetCacheInstance()
	err := p.cache.InitCache(p.ourSharedDirectory, p.ctx, p.peerHost)
	return err
}

func (p *p2pNode) initCli() {
	p.cli = cli.NewCliManager()
	p.cli.InitCliManager()
}

func (p *p2pNode) pollingNodeJoinListener() {
	var err error
	p.nodeDiscoveryChan, err = initMDNS(p.peerHost, p.RendezvousString)
	if err != nil {
		log.Printf("Start nodeJoinListener error! %s\n", err)
		p.runtimeErrChan <- err
	} else {
		for {
			peer := <-p.nodeDiscoveryChan
			// Add to node list
			p.cache.AddOnlineNode(peer)
		}
	}
}

func (p *p2pNode) pollingStdinCommandListener() {
	p.stdReader = bufio.NewReader(os.Stdin)

	for {
		//fmt.Printf("> ")
		command, err := p.stdReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from stdin. %s\n", err)
			p.runtimeErrChan <- err
		} else {
			command = strings.TrimRight(command, "\n")
			p.commandChan <- command
		}
	}
}

func (p *p2pNode) startCommandExecutor() {
	for command := range p.commandChan {
		commands := strings.Split(command, " ")
		if len(commands) >= 2 {
			p.cli.Execute(commands, p.handlerManager.GetSenderHandler(commands[1]))
		} else {
			log.Println("Missing parameters")
		}
	}
}

// Not graceful
func (p *p2pNode) bindReceiverHandler() {
	p.handlerManager = handler.NewHandlerManager()
	p.handlerManager.InitHandlerManager()
	for _, handler := range p.handlerManager.GetHandlers() {
		p.peerHost.SetStreamHandler(protocol.ID(handler.GetProtocolID()), handler.HandleReceivedStream)
	}
}

// Leave Not graceful
func (p *p2pNode) Leave() {
	commands := make([]string, 0)
	commands = append(commands, "peer", "leave")
	senderHandler := p.handlerManager.GetSenderHandler("leave")
	errs := senderHandler.OpenStreamAndSendRequest(commands)
	if len(errs) != 0 {
		log.Printf("Some errors occurred while executing %s\n", commands)
	}
}
