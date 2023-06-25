package handler

import (
	"bufio"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"log"
	"os"
	"sync"
)

type LeaveHandler struct {
	protocolID  string
	cache       *runtime.Cache
	leaveMarker []byte
}

func NewLeaveHandler() *LeaveHandler {
	return &LeaveHandler{}
}

func (l *LeaveHandler) initHandler(protocolID string) {
	l.protocolID = protocolID
	l.leaveMarker = []byte("BYE")
	l.cache = runtime.GetCacheInstance()
}

func (l *LeaveHandler) GetProtocolID() string {
	return l.protocolID
}

func (l *LeaveHandler) HandleReceivedStream(stream network.Stream) {
	l.cache.RemoveOfflineNode(stream.Conn().RemotePeer().String())
	err := stream.Close()
	if err != nil {
		log.Println("Error closing stream:", err)
	} else {
		//log.Println("Closing stream")
	}
}

func (l *LeaveHandler) OpenStreamAndSendRequest(host host.Host, queryInfos []string) []error {
	var errs []error
	var stream network.Stream
	var offlineNodes []string
	queryNodes := l.cache.GetOnlineNodes()
	wg := sync.WaitGroup{}
	for _, p := range queryNodes {
		var err error
		if err = host.Connect(l.cache.GetContext(), p); err != nil {
			log.Printf("Connection failed:failed to dial %s", p.ID.String())
			offlineNodes = append(offlineNodes, p.ID.String())
			errs = append(errs, err)
			continue
		}

		// Open a stream, this stream will be handled by HandleReceivedStream on the other end
		stream, err = host.NewStream(l.cache.GetContext(), p.ID, protocol.ID(l.GetProtocolID()))
		if err != nil {
			errs = append(errs, err)
			log.Printf("Stream open failed:%s", err)
		} else {
			wg.Add(1)
			go func(stream network.Stream) {
				defer wg.Done()
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				l.writeData(rw, l.leaveMarker)

				sErr := stream.Close()
				if sErr != nil {
					errs = append(errs, sErr)
					log.Println("Error closing stream:", sErr)
				} else {
					//log.Println("Closing stream")
				}
			}(stream)
		}
	}
	wg.Wait()
	log.Println("Node leave gracefully")
	os.Exit(0)
	return errs
}

func (l *LeaveHandler) writeData(rw *bufio.ReadWriter, sendData []byte) {
	_, err := rw.Write(sendData)
	if err != nil {
		log.Printf("Error writing to buffer:%s", err)
	} else {
		err = rw.Flush()
		if err != nil {
			log.Printf("Error flushing buffer:%s", err)
		}
	}
}
