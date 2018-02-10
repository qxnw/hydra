package http

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"
)

type apiServerAdapter struct {
}

func (h *apiServerAdapter) Resolve(c servers.IRegistryEngine, conf conf.Conf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewApiResponsiveServer(c, conf, log)
}

type webServerAdapter struct {
}

func (h *webServerAdapter) Resolve(c servers.IRegistryEngine, conf conf.Conf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewWebResponsiveServer(c, conf, log)
}

func init() {
	servers.Register("api", &apiServerAdapter{})
	servers.Register("web", &webServerAdapter{})
}
