package service

import (
	"fmt"

	"github.com/qxnw/hydra/registry"

	"github.com/qxnw/lib4go/logger"
)

//IService 服务注册组件
type IService interface {
	registry.Registry
	RegisterService(serviceName string, endPointName string, data string) (string, error)
	RegisterSeqNode(path string, data string) (string, error)
	Unregister(path string) error
}

//IServiceResolver 定义配置文件转换方法
type IServiceResolver interface {
	Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (IService, error)
}

var registers = make(map[string]IServiceResolver)

//Register 注册服务注册解析器
func Register(name string, resolver IServiceResolver) {
	if resolver == nil {
		panic("registry: Register adapter is nil")
	}
	if _, ok := registers[name]; ok {
		panic("registry: Register called twice for adapter " + name)
	}
	registers[name] = resolver
}

//NewRegister 根据适配器名称及参数返回配置处理器
func NewRegister(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (IService, error) {
	resolver, ok := registers[adapter]
	if !ok {
		return nil, fmt.Errorf("registry: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, serverName, log, servers, cross)
}
