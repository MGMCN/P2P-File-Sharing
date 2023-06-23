package cli

import (
	"fmt"
	"github.com/MGMCN/P2PFileSharing/runtime"
	"log"
)

type RuntimeChecker struct {
	cache *runtime.Cache
}

func NewRuntimeChecker() *RuntimeChecker {
	return &RuntimeChecker{}
}

func (r *RuntimeChecker) InitRuntimeChecker() {
	r.cache = runtime.GetCacheInstance()
}

func (r *RuntimeChecker) ExecuteCommand(commands []string) {
	if len(commands) == 1 {
		log.Println("Missing command")
	} else {
		switch commands[1] {
		case "list":
			ourSharedResources := r.cache.GetSharedResourcesFromCache()
			OthersSharedResourcesMap := r.cache.GetOthersSharedResourcesPeerIDList()
			resources := ""
			for _, resource := range ourSharedResources {
				formatData := fmt.Sprintf(" | %s ( %d bytes )", resource.FileName, resource.FileSize)
				resources += formatData
			}
			log.Printf("We share the following resources:%s\n", resources)
			log.Printf("%-20s | %-14s | %s\n", "Resource", "Size", "Peers")
			for _, othersSharedResourcesInfo := range OthersSharedResourcesMap {
				fsize := fmt.Sprintf("%d bytes", othersSharedResourcesInfo.SharedFileInfo.FileSize)
				log.Printf("%-20s | %-14s | %s\n", othersSharedResourcesInfo.SharedFileInfo.FileName, fsize, othersSharedResourcesInfo.SharedPeers)
			}

		default:
		}
	}
}
