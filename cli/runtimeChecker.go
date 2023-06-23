package cli

import (
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
				resources += " | " + resource
			}
			log.Printf("We share the following resources:%s\n", resources)
			log.Printf("%-15s | %s\n", "Resource", "Peers")
			for resourceName, peerIDList := range OthersSharedResourcesMap {
				log.Printf("%-15s | %s\n", resourceName, peerIDList)
			}
		default:
		}
	}
}
