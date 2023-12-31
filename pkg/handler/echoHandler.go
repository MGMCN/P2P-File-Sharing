package handler

import (
	"bufio"
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"log"
	"strings"
	"sync"
)

type EchoHandler struct {
	protocolID string
	cache      *runtime.Cache
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (e *EchoHandler) initHandler(protocolID string) {
	e.protocolID = protocolID
	e.cache = runtime.GetCacheInstance()
}

func (e *EchoHandler) GetProtocolID() string {
	return e.protocolID
}

func (e *EchoHandler) HandleReceivedStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	// Wait for read and write to complete before closing the stream
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		e.readData(rw)
	}()
	go func() {
		defer wg.Done()
		e.writeData(rw, fmt.Sprintf("Hello from %s\n", stream.Conn().LocalPeer()))
	}()
	wg.Wait()

	err := stream.Close()
	if err != nil {
		log.Printf("Error closing stream:%s", err)
	} else {
		//log.Println("Closing stream")
	}
}

func (e *EchoHandler) OpenStreamAndSendRequest(queryInfos []string) []error {
	var errs []error
	var stream network.Stream
	var offlineNodes []string
	host := e.cache.GetHost()
	queryNodes := e.cache.GetOnlineNodes()
	for _, p := range queryNodes {
		var err error
		if err = host.Connect(e.cache.GetContext(), p); err != nil {
			log.Printf("Connection failed:failed to dial %s", p.ID.String())
			offlineNodes = append(offlineNodes, p.ID.String())
			errs = append(errs, err)
			continue
		}

		// Open a stream, this stream will be handled by HandleReceivedStream on the other end
		stream, err = host.NewStream(e.cache.GetContext(), p.ID, protocol.ID(e.GetProtocolID()))
		go func(stream network.Stream) {
			if err != nil {
				errs = append(errs, err)
				log.Printf("Stream open failed:%s", err)
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				wg := sync.WaitGroup{}
				wg.Add(2)
				go func() {
					defer wg.Done()
					e.writeData(rw, fmt.Sprintf("Hello from %s\n", stream.Conn().LocalPeer()))
				}()
				go func() {
					defer wg.Done()
					e.readData(rw)
				}()
				wg.Wait()

				err = stream.Close()
				if err != nil {
					errs = append(errs, err)
					log.Println("Error closing stream:", err)
				} else {
					//log.Println("Closing stream")
				}
			}
		}(stream)
	}
	e.cache.RemoveOfflineNodes(offlineNodes)
	return errs
}

func (e *EchoHandler) readData(rw *bufio.ReadWriter) {
	str, err := rw.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from buffer:%s\n", err)
	} else {
		// Green console colour: 	\x1b[32m
		// Reset console colour: 	\x1b[0m
		log.Printf("\x1b[32m%s\x1b[0m", strings.TrimRight(str, "\n"))
	}
}

func (e *EchoHandler) writeData(rw *bufio.ReadWriter, sendData string) {
	_, err := rw.WriteString(fmt.Sprintf("%s\n", sendData))
	if err != nil {
		log.Printf("Error writing to buffer:%s", err)
	} else {
		err = rw.Flush()
		if err != nil {
			log.Printf("Error flushing buffer:%s", err)
		}
	}
}
