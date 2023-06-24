package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/MGMCN/P2PFileSharing/cli"
	"github.com/MGMCN/P2PFileSharing/handler"
	"github.com/MGMCN/P2PFileSharing/runtime"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"strings"
	"sync"
)

type p2pNode struct {
	ctx                context.Context
	peerHost           host.Host
	RendezvousString   string
	listenHost         string
	listenPort         int
	nodeDiscoveryChan  chan peer.AddrInfo
	onlineNodes        []peer.AddrInfo
	handlerManager     *handler.Manager
	stdReader          *bufio.Reader
	commandChan        chan string
	runtimeErrChan     chan error
	mutex              *sync.Mutex
	ourSharedDirectory string
	cache              *runtime.Cache
	cli                *cli.RuntimeChecker
}

func Newp2pNode() *p2pNode {
	return &p2pNode{}
}

func (p *p2pNode) InitP2PNode(ctx context.Context, RendezvousString string, listenHost string, listenPort int, ourSharedDirectory string, runtimeErrChan chan error) error {
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
	p.mutex = new(sync.Mutex)
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
	return err
}

func (p *p2pNode) initCacheStorage() error {
	p.cache = runtime.GetCacheInstance()
	err := p.cache.InitCache(p.ourSharedDirectory, p.ctx)
	return err
}

func (p *p2pNode) initCli() {
	p.cli = cli.NewRuntimeChecker()
	p.cli.InitRuntimeChecker()
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
			//log.Printf("%s\n", peer)
			// Add to node list
			p.mutex.Lock()
			p.onlineNodes = append(p.onlineNodes, peer)
			p.mutex.Unlock()
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
			//log.Printf("%s\n", command)
			command = strings.TrimRight(command, "\n")
			p.commandChan <- command
		}
	}
}

func (p *p2pNode) startCommandExecutor() {
	for command := range p.commandChan {
		commands := strings.Split(command, " ")
		if commands[0] != "" {
			if commands[0] == "cache" {
				p.cli.ExecuteCommand(commands)
			} else {
				// Not graceful
				if commands[0] == "peer" && len(commands) >= 2 {
					senderHandler := p.handlerManager.GetSenderHandler(commands[1])
					if senderHandler != nil {
						p.mutex.Lock()
						tempOnlineNodes := p.onlineNodes
						p.mutex.Unlock()
						go func(tempOnlineNodes []peer.AddrInfo) {
							// Should we move peerHost and onlineNodes to cache also
							errs, offlineNodesID := senderHandler.OpenStreamAndSendRequest(p.peerHost, tempOnlineNodes, commands)
							if len(errs) != 0 {
								log.Printf("Some errors occurred while executing %s\n", commands)
							}
							if len(offlineNodesID) != 0 {
								p.removeOfflineNodes(offlineNodesID)
								p.removeOfflineNodesResources(offlineNodesID)
							}
						}(tempOnlineNodes)
					}
				} else {
					log.Println("Missing parameters")
				}

			}
		}
	}
}

func (p *p2pNode) removeOfflineNodesResources(offlineNodesID []string) {
	p.cache.RemoveOfflineNodesSharedResources(offlineNodesID)
}

func (p *p2pNode) removeOfflineNodes(offlineNodesID []string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var updatedOnlineNodes []peer.AddrInfo
	for _, onlineNode := range p.onlineNodes {
		found := false
		for _, offlineNodeID := range offlineNodesID {
			if offlineNodeID == onlineNode.ID.String() {
				log.Printf("Found %s offline\n", offlineNodeID)
				found = true
				break
			}
		}
		if !found {
			updatedOnlineNodes = append(updatedOnlineNodes, onlineNode)
		}
	}
	p.onlineNodes = updatedOnlineNodes
}

// Not graceful
func (p *p2pNode) bindReceiverHandler() {
	p.handlerManager = handler.NewHandlerManager()
	p.handlerManager.InitHandlerManager()
	for _, handler := range p.handlerManager.GetHandlers() {
		p.peerHost.SetStreamHandler(protocol.ID(handler.GetProtocolID()), handler.HandleReceivedStream)
	}
}
