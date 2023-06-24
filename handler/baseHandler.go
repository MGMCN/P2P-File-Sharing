package handler

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type BaseStreamHandler interface {
	initHandler(protocolID string)
	GetProtocolID() string
	HandleReceivedStream(stream network.Stream)
	OpenStreamAndSendRequest(host host.Host, queryNodes []peer.AddrInfo, queryInfos []string) ([]error, []string)
}
