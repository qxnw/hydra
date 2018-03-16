package http

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"
)

type apiServerAdapter struct {
}

func (h *apiServerAdapter) Resolve(registryAddr string, conf conf.IServerConf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewApiResponsiveServer(registryAddr, conf, log)
}

type webServerAdapter struct {
}

func (h *webServerAdapter) Resolve(registryAddr string, conf conf.IServerConf, log *logger.Logger) (servers.IRegistryServer, error) {
	return NewWebResponsiveServer(registryAddr, conf, log)
}

func init() {
	servers.Register("api", &apiServerAdapter{})
	servers.Register("web", &webServerAdapter{})
}
