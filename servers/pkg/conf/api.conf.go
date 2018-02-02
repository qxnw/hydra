package conf

import "github.com/qxnw/hydra/conf"

type ApiServerConf struct {
	*ServerConf
	OnlyAllowAjaxRequest bool
}

func NewApiServerConf(domain string, name string, typeName string, tag string, registryNode string, mask string, timeout int) *ApiServerConf {
	return &ApiServerConf{ServerConf: NewConf(domain, name, typeName, tag, registryNode, mask, timeout)}
}

func NewApiServerConfBy(nconf conf.Conf) *ApiServerConf {
	return &ApiServerConf{
		ServerConf: NewConfBy(nconf),
	}
}
