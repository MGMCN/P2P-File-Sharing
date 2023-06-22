package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/MGMCN/P2PFileSharing/runtime"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"log"
	"sync"
)

type SearchHandler struct {
	protocolID string
	cache      *runtime.Cache
	endMarker  []byte
}

type queryInfo struct {
	Command string `json:"command"`
	Keyword string `json:"keyword"`
}

type sharedInfo struct {
	Id        string   `json:"id"`
	Resources []string `json:"resources"`
}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

func (s *SearchHandler) initHandler(protocolID string) {
	s.protocolID = protocolID
	s.endMarker = []byte("END")
	s.cache = runtime.GetCacheInstance()
}

func (s *SearchHandler) GetProtocolID() string {
	return s.protocolID
}

func (s *SearchHandler) HandleReceivedStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.readData(rw, true)
	}()
	go func() {
		defer wg.Done()
		var infos sharedInfo
		resources := s.cache.GetSharedResourcesFromCache()
		infos.Resources = resources
		infos.Id = stream.Conn().LocalPeer().String()
		jsonData, err := json.Marshal(infos)
		if err != nil {
			log.Printf("json.Marshal error:%s", err)
		} else {
			s.writeData(rw, jsonData)
			s.writeData(rw, s.endMarker)
		}
	}()
	wg.Wait()

	err := stream.Close()
	if err != nil {
		log.Println("Error closing stream:", err)
	} else {
		//log.Println("Closing stream")
	}
}

func (s *SearchHandler) SendRequest(ctx context.Context, host host.Host, queryNodes []peer.AddrInfo, queryInfos []string) ([]error, []string) {
	var errs []error
	var stream network.Stream
	var infos queryInfo
	var offlineNodes []string
	if len(queryInfos) == 1 {
		infos = queryInfo{
			Command: "search",
			Keyword: "all",
		}
	} else if len(queryInfos) > 1 {
		infos = queryInfo{
			Command: "search",
			Keyword: queryInfos[0],
		}
	}

	jsonData, err := json.Marshal(infos)
	if err != nil {
		errs = append(errs, err)
		log.Printf("json.Marshal error:%s", err)
	} else {
		for _, p := range queryNodes {
			//log.Println("Try connect -> ", p)
			if err = host.Connect(ctx, p); err != nil {
				log.Printf("Connection failed:failed to dial %s", p.ID.String())
				offlineNodes = append(offlineNodes, p.ID.String())
				errs = append(errs, err)
				continue
			}

			// Open a stream, this stream will be handled by HandleReceivedStream on the other end
			stream, err = host.NewStream(ctx, p.ID, protocol.ID(s.GetProtocolID()))
			if err != nil {
				errs = append(errs, err)
				log.Printf("Stream open failed:%s", err)
			} else {
				go func(stream network.Stream) {
					rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

					wg := sync.WaitGroup{}
					wg.Add(2)
					go func() {
						defer wg.Done()
						s.writeData(rw, jsonData)
						s.writeData(rw, s.endMarker)
					}()
					go func() {
						defer wg.Done()
						s.readData(rw, false)
					}()
					wg.Wait()

					err = stream.Close()
					if err != nil {
						errs = append(errs, err)
						log.Println("Error closing stream:", err)
					} else {
						//log.Println("Closing stream")
					}
				}(stream)
			}
		}
	}

	return errs, offlineNodes
}

func (s *SearchHandler) readData(rw *bufio.ReadWriter, received bool) {
	var jsonData []byte
	var queryInfos queryInfo
	var sharedInfos sharedInfo
	var err error
	var n int
	var endFlag = false
	buffer := make([]byte, 1024)

	//n, err = io.ReadFull(rw, buffer)
	//jsonData = append(jsonData, buffer[:n]...)

	for {
		n, err = rw.Read(buffer)
		if err != nil {
			break
		}
		if bytes.Equal(buffer[:n], s.endMarker) {
			endFlag = true
			break
		}
		jsonData = append(jsonData, buffer[:n]...)
	}

	if !endFlag {
		log.Printf("Error reading from buffer:%s\n", err)
	} else {
		if received {
			err = json.Unmarshal(jsonData, &queryInfos)
		} else {
			err = json.Unmarshal(jsonData, &sharedInfos)
		}
		if err != nil {
			log.Printf("json.Unmarshal error:%s\n", err)
		} else {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			if received {
				// do something
				//log.Printf("\x1b[32m%s\x1b[0m", queryInfos.Keyword)
			} else {
				log.Printf("\x1b[32mUpdateOthersSharedResources from %s\x1b[0m", sharedInfos.Id)
				s.cache.UpdateOthersSharedResources(sharedInfos.Resources, sharedInfos.Id)
				// test
				//log.Println(s.cache.GetOthersSharedResourcesPeerIDListFilterByResourceName("a.txt"))
				//log.Println(s.cache.GetOthersSharedResourcesPeerIDListFilterByResourceName("b.txt"))
				//log.Println(s.cache.GetOthersSharedResourcesPeerIDListFilterByResourceName("c.txt"))
			}

		}
	}
}

func (s *SearchHandler) writeData(rw *bufio.ReadWriter, sendData []byte) {
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
