package rpc

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"
)

type rpcServerAdapter struct {
}

func (h *rpcServerAdapter) Resolve(c servers.IRegistryEngine, conf conf.Conf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewRpcResponsiveServer(c, conf, log)
}

func init() {
	servers.Register("rpc", &rpcServerAdapter{})
}
