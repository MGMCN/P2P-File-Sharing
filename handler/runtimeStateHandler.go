package handler

import (
	"context"
	"github.com/MGMCN/P2PFileSharing/storage"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
)

type StateHandler struct {
	protocolID string
	cache      *storage.Cache
}

func NewStateHandler() *StateHandler {
	return &StateHandler{}
}

func (s *StateHandler) initHandler(protocolID string) {
	s.protocolID = protocolID
	s.cache = storage.GetCacheInstance()
}

func (s *StateHandler) GetProtocolID() string {
	return s.protocolID
}

func (s *StateHandler) HandleReceivedStream(stream network.Stream) {
}

func (s *StateHandler) SendRequest(ctx context.Context, host host.Host, queryNodes []peer.AddrInfo, queryInfos []string) (error, []string) {
	var err error
	var offlineNodes []string
	if len(queryInfos) == 0 {
		log.Println("Missing parameters")
	}
	switch queryInfos[0] {
	case "list":
		ourSharedResources := s.cache.GetSharedResourcesFromCache()
		OthersSharedResourcesMap := s.cache.GetOthersSharedResourcesPeerIDList()
		resources := ""
		for _, resource := range ourSharedResources {
			resources += " | " + resource
		}
		log.Printf("We share the following resources:%s\n", resources)
		log.Printf("%-15s | %s\n", "Resource", "Peers")
		for resourceName, peerIDList := range OthersSharedResourcesMap {
			log.Printf("%-15s | %s\n", resourceName, peerIDList)
		}
	default:
	}
	return err, offlineNodes
}
