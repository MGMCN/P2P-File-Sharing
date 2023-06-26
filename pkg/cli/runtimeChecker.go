package cli

import (
	"fmt"
	"github.com/MGMCN/P2PFileSharing/pkg/runtime"
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
		// should put here ?
		log.Println("Missing parameters")
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
			log.Printf("The resources shared by other nodes are listed in the table below\n")
			log.Printf("%-30s | %-20s | %s\n", "Resource", "Size", "Peers")
			for _, othersSharedResourcesInfo := range OthersSharedResourcesMap {
				fsize := fmt.Sprintf("%d bytes", othersSharedResourcesInfo.SharedFileInfo.FileSize)
				log.Printf("%-30s | %-20s | %s\n", othersSharedResourcesInfo.SharedFileInfo.FileName, fsize, othersSharedResourcesInfo.SharedPeers)
			}

		default:
		}
	}
}
