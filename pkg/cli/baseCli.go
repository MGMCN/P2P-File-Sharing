package cli

import "github.com/MGMCN/P2PFileSharing/pkg/handler"

type BaseCli interface {
	Execute(commands []string, handler handler.BaseStreamHandler)
}
