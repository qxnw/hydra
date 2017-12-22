package engine

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//IServerContext 服务器上下文
type IContainer interface {
	context.RPCInvoker
	IVarInvoker
	GetDomainName() string
	GetServerName() string
}
type IVarInvoker interface {
	GetVarParam(tp string, name string) (string, error)
}
type Container struct {
	context.RPCInvoker
	registry   registry.Registry
	domain     string
	serverName string
}

//NewServerContext 服务器上下文
func NewContainer(rpci *rpc.Invoker, registryAddr string, domain string, serverName string, logg logger.ILogger) *Container {
	registry, _ := registry.NewRegistryWithAddress(registryAddr, logg)
	return &Container{RPCInvoker: rpci, registry: registry, domain: domain, serverName: serverName}
}

//GetDomainName 获取域信息
func (c *Container) GetDomainName() string {
	return c.domain
}

//GetServerName 获取服务器名称
func (c *Container) GetServerName() string {
	return c.serverName
}

//GetVarParam 获取配置参数
func (c *Container) GetVarParam(tp string, name string) (string, error) {
	buff, _, err := c.registry.GetValue(fmt.Sprintf("/%s/var/%s/%s", c.domain, tp, name))
	if err != nil {
		return "", err
	}
	return string(buff), nil
}
