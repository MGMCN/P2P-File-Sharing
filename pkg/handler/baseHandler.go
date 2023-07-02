package handler

import (
	"github.com/libp2p/go-libp2p/core/network"
)

type BaseStreamHandler interface {
	initHandler(protocolID string)
	GetProtocolID() string
	HandleReceivedStream(stream network.Stream)
	OpenStreamAndSendRequest(queryInfos []string) []error
}
