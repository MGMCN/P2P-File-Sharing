package storage

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	ourSharedDirectory    string
	ourSharedResources    []string
	othersSharedResources map[string][]string // Resource name -> Peer list
	mutex                 *sync.Mutex
	oMutex                *sync.Mutex
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

func (c *Cache) InitCache(ourSharedDirectory string) error {
	c.ourSharedDirectory = ourSharedDirectory
	c.mutex = new(sync.Mutex)
	c.oMutex = new(sync.Mutex)
	c.othersSharedResources = make(map[string][]string)
	err := c.ReadSharedResourcesIntoCache()
	if err != nil {
		log.Printf("ReadSharedResourcesIntoCache error:%s\n", err)
	}
	return err
}

func (c *Cache) RemoveOfflineNodesSharedResources(peerIDList []string) {

}

func (c *Cache) UpdateOthersSharedResources(resources []string, peerID string) {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()

	for _, resource := range resources {
		found := false
		// We should use set
		for _, storedPeerID := range c.othersSharedResources[resource] {
			if storedPeerID == peerID {
				found = true
			}
		}
		if !found {
			c.othersSharedResources[resource] = append(c.othersSharedResources[resource], peerID)
		}
	}
}

func (c *Cache) GetOthersSharedResourcesPeerIDList() map[string][]string {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()
	return c.othersSharedResources
}

func (c *Cache) GetOthersSharedResourcesPeerIDListFilterByResourceName(resourceName string) []string {
	c.oMutex.Lock()
	defer c.oMutex.Unlock()

	return c.othersSharedResources[resourceName]
}

func (c *Cache) GetSharedResourcesFromCache() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.ourSharedResources
}

func (c *Cache) AddDownloadedResource(resourceName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ourSharedResources = append(c.ourSharedResources, resourceName)
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
			filename := filepath.Base(path)
			c.ourSharedResources = append(c.ourSharedResources, filename)
		}

		return nil
	})
	if err != nil {
		log.Printf("Error traversing local folders:%s\n", err)
	}
	return err
}
