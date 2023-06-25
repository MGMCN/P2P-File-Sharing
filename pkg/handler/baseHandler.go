package handler

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

type BaseStreamHandler interface {
	initHandler(protocolID string)
	GetProtocolID() string
	HandleReceivedStream(stream network.Stream)
	OpenStreamAndSendRequest(host host.Host, queryInfos []string) []error
}
