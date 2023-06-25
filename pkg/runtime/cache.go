package runtime

import (
	"context"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type FileInfo struct {
	FileName string
	FileSize int64
}

type OtherSharedFileInfo struct {
	SharedFileInfo FileInfo
	SharedPeers    []string
}

type Cache struct {
	onlineNodes           []peer.AddrInfo
	ourSharedDirectory    string
	ourSharedResources    []FileInfo
	othersSharedResources map[string]OtherSharedFileInfo // Resource name -> Peer list
	ctx                   context.Context
	mutex                 *sync.Mutex
	oMutex                *sync.Mutex
	nMutex                *sync.Mutex
}

// Not graceful
var (
	instance *Cache
	once     sync.Once
)

func GetCacheInstance() *Cache {
	once.Do(func() {
		instance = &Cache{}
	})
	return instance
}

func (c *Cache) InitCache(ourSharedDirectory string, ctx context.Context) error {
	c.ourSharedDirectory = ourSharedDirectory
	c.ctx = ctx
	c.mutex = new(sync.Mutex)
	c.oMutex = new(sync.Mutex)
	c.nMutex = new(sync.Mutex)
	c.othersSharedResources = make(map[string]OtherSharedFileInfo)
	err := c.ReadSharedResourcesIntoCache()
	if err != nil {
		log.Printf("ReadSharedResourcesIntoCache error:%s\n", err)
	}
	return err
}

func (c *Cache) AddOnlineNode(p peer.AddrInfo) {
	c.nMutex.Lock()
	defer c.nMutex.Unlock()

	c.onlineNodes = append(c.onlineNodes, p)
}

func (c *Cache) GetOnlineNodes() []peer.AddrInfo {
	c.nMutex.Lock()
	defer c.nMutex.Unlock()

	return c.onlineNodes
}

func (c *Cache) RemoveOfflineNode(offlineNodeID string) {
	c.nMutex.Lock()
	defer c.nMutex.Unlock()

	var found = false
	var foundIndex = 0
	for _, onlineNode := range c.onlineNodes {
		if offlineNodeID == onlineNode.ID.String() {
			log.Printf("Found %s offline\n", offlineNodeID)
			found = true
			break
		}
		foundIndex += 1
	}
	if found {
		c.onlineNodes = append(c.onlineNodes[:foundIndex], c.onlineNodes[foundIndex+1:]...)
	}
}

func (c *Cache) RemoveOfflineNodes(offlineNodesID []string) {
	c.nMutex.Lock()
	var updatedOnlineNodes []peer.AddrInfo
	for _, onlineNode := range c.onlineNodes {
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
	c.onlineNodes = updatedOnlineNodes
	c.nMutex.Unlock()

	c.removeOfflineNodesSharedResources(offlineNodesID)
}

func (c *Cache) ReadSharedResourcesIntoCache() error {
	var err error
	err = c.traversingResourceFolder()
	return err
}

func (c *Cache) traversingResourceFolder() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := filepath.Walk(c.ourSharedDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			resourceFile := FileInfo{
				FileName: filepath.Base(path),
				FileSize: info.Size(),
			}
			c.ourSharedResources = append(c.ourSharedResources, resourceFile)
		}

		return nil
	})
	if err != nil {
		log.Printf("Error traversing local folders:%s\n", err)
	}
	return err
}

func (c *Cache) removeOfflineNodesSharedResources(peerIDList []string) {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()

	for _, offlinePeerID := range peerIDList {
		for resourceName, othersSharedResourcesInfo := range c.othersSharedResources {
			var foundIndex int
			var found = false
			for index, onlinePeerID := range othersSharedResourcesInfo.SharedPeers {
				if offlinePeerID == onlinePeerID {
					found = true
					foundIndex = index
					break
				}
			}
			if found {
				othersSharedResourcesInfo.SharedPeers = append(othersSharedResourcesInfo.SharedPeers[:foundIndex], othersSharedResourcesInfo.SharedPeers[foundIndex+1:]...)
				c.othersSharedResources[resourceName] = othersSharedResourcesInfo
			}
		}
	}
}

func (c *Cache) UpdateOthersSharedResources(resources []FileInfo, peerID string) {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()
	for _, resource := range resources {
		found := false
		// We should use set
		othersSharedResourcesInfo := c.othersSharedResources[resource.FileName]
		for _, storedPeerID := range othersSharedResourcesInfo.SharedPeers {
			if storedPeerID == peerID {
				found = true
				break
			}
		}
		if !found {
			othersSharedResourcesInfo.SharedFileInfo.FileName = resource.FileName
			othersSharedResourcesInfo.SharedFileInfo.FileSize = resource.FileSize
			//
			othersSharedResourcesInfo.SharedPeers = append(othersSharedResourcesInfo.SharedPeers, peerID)
			c.othersSharedResources[resource.FileName] = othersSharedResourcesInfo
		}
	}
}

func (c *Cache) AddDownloadedResource(resourceName string, resourceSize int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	resourceFile := FileInfo{
		FileName: resourceName,
		FileSize: resourceSize,
	}

	c.ourSharedResources = append(c.ourSharedResources, resourceFile)
}

func (c *Cache) GetContext() context.Context {
	return c.ctx
}

func (c *Cache) GetOurSharedDirectory() string {
	return c.ourSharedDirectory
}

func (c *Cache) GetOthersSharedResourcesPeerIDList() map[string]OtherSharedFileInfo {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()
	return c.othersSharedResources
}

func (c *Cache) GetOthersSharedResourcesInfosFilterByResourceName(resourceName string) OtherSharedFileInfo {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()

	return c.othersSharedResources[resourceName]
}

func (c *Cache) GetSharedResourcesFromCache() []FileInfo {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.ourSharedResources
}
