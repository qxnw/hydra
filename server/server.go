package server

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
)

//IHydraServer 服务器接口，可通过单个变量和配置文件方式设置服务器启用参数
type IHydraServer interface {
	Notify(registry.Conf) error
	GetAddress() string
	Start() error
	Shutdown()
}

//IServerAdapter 服务解析器
type IServerAdapter interface {
	Resolve(c context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (IHydraServer, error)
}

var serverResolvers = make(map[string]IServerAdapter)

//Register 注册服务适配器
func Register(name string, resolver IServerAdapter) {
	if _, ok := serverResolvers[name]; ok {
		panic("server: Register called twice for adapter " + name)
	}
	serverResolvers[name] = resolver
}

//NewServer 根据适配器名称生成服务
func NewServer(adapter string, c context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (IHydraServer, error) {
	if resolver, ok := serverResolvers[adapter]; ok {
		return resolver.Resolve(c, r, conf, logger)
	}
	return nil, fmt.Errorf("server: unknown adapter name %q (forgotten import?)", adapter)

}
