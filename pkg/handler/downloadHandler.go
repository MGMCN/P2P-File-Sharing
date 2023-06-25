package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"io"
	"log"
	"os"
	"sync"
)

type DownloadHandler struct {
	protocolID string
	cache      *runtime.Cache
	endMarker  []byte
}

type queryResources struct {
	FileName    string
	StartOffset int64
	ReadSize    int64
}

func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{}
}

func (d *DownloadHandler) initHandler(protocolID string) {
	d.protocolID = protocolID
	d.endMarker = []byte("END")
	d.cache = runtime.GetCacheInstance()
}

func (d *DownloadHandler) GetProtocolID() string {
	return d.protocolID
}

func (d *DownloadHandler) HandleReceivedStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	err, queryResourcesData := d.readQueryData(rw)
	log.Printf("Someone want resource %s\n", queryResourcesData.FileName)
	if err != nil {
		log.Println("readQueryData error")
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.readFileAndWriteToStream(rw, queryResourcesData)
		}()
		wg.Wait()

		err = stream.Close()
		if err != nil {
			log.Println("Error closing stream:", err)
		} else {
			//log.Println("Closing stream")
		}
	}
}

func (d *DownloadHandler) OpenStreamAndSendRequest(host host.Host, queryInfos []string) []error {
	var errs []error
	var stream network.Stream
	var offlineNodes []string
	var jsonData []byte
	var err error
	queryNodes := d.cache.GetOnlineNodes()
	if len(queryInfos) < 3 {
		log.Println("Missing parameters")
	} else {
		queryFileName := queryInfos[2]
		othersSharedResourcesInfos := d.cache.GetOthersSharedResourcesInfosFilterByResourceName(queryFileName)
		sharedPeersID := othersSharedResourcesInfos.SharedPeers
		lenOfSharedPeersIDList := len(sharedPeersID)
		if lenOfSharedPeersIDList == 0 {
			log.Println("No one has the resources we want")
		} else {
			fileSize := othersSharedResourcesInfos.SharedFileInfo.FileSize
			fileChunkSize := fileSize / int64(lenOfSharedPeersIDList)
			log.Printf("Caculated fileChunkSize:%d bytes", fileChunkSize)

			downloadFinishWG := sync.WaitGroup{}
			var startOffset int64 = 0
			for index, peerID := range sharedPeersID {
				if index == lenOfSharedPeersIDList-1 {
					// last time read all
					fileChunkSize = fileSize - startOffset
				}
				infos := queryResources{
					FileName:    queryFileName,
					StartOffset: startOffset,
					ReadSize:    fileChunkSize,
				}
				//log.Printf("Request chunk info: %s\n", infos)
				jsonData, err = json.Marshal(infos)
				if err != nil {
					errs = append(errs, err)
					log.Printf("json.Marshal error:%s", err)
				} else {
					for _, p := range queryNodes {
						if peerID == p.ID.String() {
							if err = host.Connect(d.cache.GetContext(), p); err != nil {
								log.Printf("Connection failed:failed to dial %s", p.ID.String())
								offlineNodes = append(offlineNodes, p.ID.String())
								errs = append(errs, err)
							}

							// Open a stream, this stream will be handled by HandleReceivedStream on the other end
							stream, err = host.NewStream(d.cache.GetContext(), p.ID, protocol.ID(d.GetProtocolID()))
							if err != nil {
								errs = append(errs, err)
								log.Printf("Stream open failed:%s", err)
							} else {
								startOffset += fileChunkSize
								downloadFinishWG.Add(1)
								go func(stream network.Stream, index int, fileName string, jsonData []byte) {
									rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
									defer downloadFinishWG.Done()

									singleDownloadFinishWG := sync.WaitGroup{}
									singleDownloadFinishWG.Add(2)
									go func() {
										defer singleDownloadFinishWG.Done()
										d.writeData(rw, jsonData)
										d.writeData(rw, d.endMarker)
									}()
									go func() {
										defer singleDownloadFinishWG.Done()
										d.readReceivedFileChunk(rw, index, fileName)
									}()
									singleDownloadFinishWG.Wait()

									sErr := stream.Close()
									if sErr != nil {
										errs = append(errs, sErr)
										log.Println("Error closing stream:", sErr)
									} else {
										//log.Println("Closing stream")
									}
									log.Printf("Received file chunk from %s", stream.Conn().RemotePeer())
								}(stream, index, queryFileName, jsonData)
							}
							break
						}
					}
				}
			}
			downloadFinishWG.Wait()

			err = d.mergeFile(queryFileName, lenOfSharedPeersIDList)
			if err != nil {
				errs = append(errs, err)
				log.Printf("Failed to merge chunk of %s", queryFileName)
			} else {
				d.cache.AddDownloadedResource(queryFileName, fileSize)
				log.Printf("Merge chunk of %s successfully", queryFileName)
			}
			d.cache.RemoveOfflineNodes(offlineNodes)
		}
	}
	return errs
}

func (d *DownloadHandler) writeData(rw *bufio.ReadWriter, sendData []byte) {
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

func (d *DownloadHandler) readReceivedFileChunk(rw *bufio.ReadWriter, index int, fileName string) {
	var err error
	var n int

	tmpFilePath := fmt.Sprintf("%s%d.tmp", d.cache.GetOurSharedDirectory()+fileName, index)
	file, err := os.Create(tmpFilePath)
	if err != nil {
		log.Printf("Create %s error\n", tmpFilePath)
		return
	}
	defer file.Close()

	for {
		buffer := make([]byte, 2048)
		n, err = rw.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		_, err = file.Write(buffer[:n])
		//log.Printf("%d jsonByteData:%s\n", index, buffer[:n])
		if err != nil {
			log.Printf("Write buffer to %s error", tmpFilePath)
			return
		}
	}
}

func (d *DownloadHandler) readQueryData(rw *bufio.ReadWriter) (error, queryResources) {
	var jsonData []byte
	var queryResourcesInfo queryResources
	var err error
	var n int
	var endFlag = false
	buffer := make([]byte, 1024)

	for {
		n, err = rw.Read(buffer)
		if err != nil {
			break
		}
		if bytes.Equal(buffer[:n], d.endMarker) {
			endFlag = true
			break
		}
		jsonData = append(jsonData, buffer[:n]...)
	}

	if !endFlag {
		log.Printf("Error reading from buffer:%s\n", err)
	} else {
		err = json.Unmarshal(jsonData, &queryResourcesInfo)
		if err != nil {
			log.Printf("json.Unmarshal error:%s\n", err)
		}
	}

	return err, queryResourcesInfo
}

func (d *DownloadHandler) readFileAndWriteToStream(rw *bufio.ReadWriter, queryResourcesInfo queryResources) {
	fileName := queryResourcesInfo.FileName
	startOffset := queryResourcesInfo.StartOffset
	readSize := queryResourcesInfo.ReadSize
	//log.Println(fileName, startOffset, readSize)

	file, err := os.Open(d.cache.GetOurSharedDirectory() + fileName)
	if err != nil {
		log.Printf("Can not open file: %s\n", err)
	} else {
		defer file.Close()

		_, err = file.Seek(startOffset, io.SeekStart)
		if err != nil {
			fmt.Printf("Can not set offset: %s\n", err)
			return
		}

		var totalBytesRead int64 = 0
		var readBytesLen int64 = 0
		var bufferSize int64 = 2048

		for totalBytesRead < readSize {

			buffer := make([]byte, bufferSize)
			//log.Println(totalBytesRead+bufferSize, readSize, readSize-totalBytesRead)
			if totalBytesRead+bufferSize > readSize {
				_, err = file.Read(buffer[:readSize-totalBytesRead])
				readBytesLen = readSize - totalBytesRead
				totalBytesRead = readSize
			} else {
				_, err = file.Read(buffer)
				readBytesLen = bufferSize
				totalBytesRead += bufferSize
			}
			if err != nil && err != io.EOF {
				log.Printf("Read file error: %s\n", err)
				return
			}

			d.writeData(rw, buffer[:readBytesLen])

			if err == io.EOF {
				//log.Println("Read file complete (io.EOF)")
				break
			} else if totalBytesRead == readSize {
				//log.Println("Read file complete (totalBytesRead == readSize)")
				break
			}
		}
	}
}

func (d *DownloadHandler) mergeFile(fileName string, tmpFileCount int) error {
	var tmpFile *os.File

	baseDir := d.cache.GetOurSharedDirectory()
	mergedFilePath := baseDir + fileName
	mergedFile, err := os.OpenFile(mergedFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Printf("Create %s error\n", mergedFilePath)
	} else {
		defer mergedFile.Close()

		for count := 0; count < tmpFileCount; count++ {
			tmpFilePath := fmt.Sprintf("%s%d.tmp", baseDir+fileName, count)
			tmpFile, err = os.Open(tmpFilePath)
			if err != nil {
				log.Printf("Can not open file: %s\n", tmpFilePath)
				break
			} else {
				_, err = io.Copy(mergedFile, tmpFile)
				tmpFile.Close()
				if err != nil {
					log.Printf("Failed to write data from %s to merged file: %s\n", tmpFilePath, mergedFilePath)
					break
				}
				err = os.Remove(tmpFilePath)
				if err != nil {
					log.Printf("Failed to remove tmp file: %s\n", tmpFilePath)
				}
			}
		}
	}
	return err
}
