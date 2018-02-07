package registry

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"
)

type serverAdapter struct {
}

func (h *serverAdapter) Resolve(c servers.IRegistryEngine, conf conf.Conf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewRegistryServer(c, conf, log)
}

func init() {
	servers.Register("mqc", &serverAdapter{})
}
