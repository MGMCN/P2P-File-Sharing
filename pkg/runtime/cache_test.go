package runtime

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"sync"
	"testing"
)

func addTestSamplesToCache() (*Cache, []string) {
	cache := GetCacheInstance()
	cache.InitCache("./", nil)
	p1 := peer.AddrInfo{ID: "peer1"}
	p2 := peer.AddrInfo{ID: "peer2"}
	expectedPeers := make([]string, 0)
	expectedPeers = append(expectedPeers, p1.ID.String())
	expectedPeers = append(expectedPeers, p2.ID.String())
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		cache.AddOnlineNode(p1)
	}()
	go func() {
		defer wg.Done()
		cache.AddOnlineNode(p2)
	}()
	wg.Wait()
	return cache, expectedPeers
}

func TestLengthOfOnlineNodesWhenDoAddOperation(t *testing.T) {
	cache, _ := addTestSamplesToCache()
	expected := 2
	result := len(cache.GetOnlineNodes())
	if result != expected {
		t.Errorf("Add peer1,peer2 returned length=%d, expected length=%d", result, expected)
	}
}

func TestDoesTheAddedNodeMatch(t *testing.T) {
	cache, expectedPeers := addTestSamplesToCache()
	onlineNodes := cache.GetOnlineNodes()
	for _, expectedPeer := range expectedPeers {
		found := false
		for _, onlineNode := range onlineNodes {
			if expectedPeer == onlineNode.ID.String() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s not found", expectedPeer)
		}
	}
}

func TestRemoveOfflineNode(t *testing.T) {
	cache, expectedPeers := addTestSamplesToCache()
	cache.RemoveOfflineNode(expectedPeers[0])
	onlineNodes := cache.GetOnlineNodes()
	if onlineNodes[0].ID.String() != expectedPeers[1] {
		t.Errorf("Expected %s not found", expectedPeers[1])
	}
}

func TestRemoveOfflineNodes(t *testing.T) {
	cache, expectedPeers := addTestSamplesToCache()
	cache.RemoveOfflineNodes(expectedPeers)
	onlineNodes := cache.GetOnlineNodes()
	returnedLen := len(onlineNodes)
	if returnedLen != 0 {
		t.Errorf("Returned length=%d expected length=0", returnedLen)
	}
}
