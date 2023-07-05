package cli

import (
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/handler"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
	"log"
)

type CacheCli struct {
	cache *runtime.Cache
}

func NewCacheCli() *CacheCli {
	return &CacheCli{}
}

func (cc *CacheCli) InitCacheCli() {
	cc.cache = runtime.GetCacheInstance()
}

func (cc *CacheCli) Execute(commands []string, handler handler.BaseStreamHandler) {
	switch commands[1] {
	case "list":
		ourSharedResources := cc.cache.GetSharedResourcesFromCache()
		othersSharedResourcesMap := cc.cache.GetOthersSharedResourcesPeerIDList()
		resources := ""
		for _, resource := range ourSharedResources {
			formatData := fmt.Sprintf(" | %s ( %d bytes )", resource.FileName, resource.FileSize)
			resources += formatData
		}
		log.Printf("We share the following resources:%s\n", resources)
		log.Printf("The resources shared by other nodes are listed in the table below\n")
		log.Printf("%-30s | %-20s | %s\n", "Resource", "Size", "Peers")
		for _, othersSharedResourcesInfo := range othersSharedResourcesMap {
			fsize := fmt.Sprintf("%d bytes", othersSharedResourcesInfo.SharedFileInfo.FileSize)
			log.Printf("%-30s | %-20s | %s\n", othersSharedResourcesInfo.SharedFileInfo.FileName, fsize, othersSharedResourcesInfo.SharedPeers)
		}

	default:
		log.Println("Please make sure you are entering a valid cache command")
	}
}
