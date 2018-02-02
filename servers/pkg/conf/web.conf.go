package conf

import (
	"github.com/qxnw/hydra/conf"
)

type WebServerConf struct {
	*ServerConf
	View *View
}

func NewWebServerConf(domain string, name string, typeName string, tag string, registryNode string, mask string, timeout int) *WebServerConf {
	return &WebServerConf{ServerConf: NewConf(domain, name, typeName, tag, registryNode, mask, timeout)}
}

func NewWebServerConfBy(nconf conf.Conf) *WebServerConf {
	return &WebServerConf{
		ServerConf: NewConfBy(nconf),
		View:       &View{Path: "../views", Left: "{{", Right: "}}"},
	}
}
