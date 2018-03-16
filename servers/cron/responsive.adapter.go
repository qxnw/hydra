package cron

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"
)

type rpcServerAdapter struct {
}

func (h *rpcServerAdapter) Resolve(c string, conf conf.IServerConf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewCronResponsiveServer(c, conf, log)
}

func init() {
	servers.Register("cron", &rpcServerAdapter{})
}
