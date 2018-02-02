package conf

import "github.com/qxnw/hydra/conf"

type RpcServerConf struct {
	*ServerConf
}

func NewRpcServerConf(domain string, name string, typeName string, tag string, registryNode string, mask string, timeout int) *RpcServerConf {
	return &RpcServerConf{ServerConf: NewConf(domain, name, typeName, tag, registryNode, mask, timeout)}
}

func NewRpcServerConfBy(nconf conf.Conf) *RpcServerConf {
	return &RpcServerConf{
		ServerConf: NewConfBy(nconf),
	}
}
