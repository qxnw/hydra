package service

import (
	"fmt"

	"github.com/qxnw/lib4go/logger"
)

//IServiceRegistry 服务注册组件
type IServiceRegistry interface {
	RegisterTempNode(serviceName string, endPointName string, data string) (string, error)
	RegisterSeqNode(path string, data string) (string, error)
	Unregister(path string) error
	Close() error
}

//IRegisterResolver 定义配置文件转换方法
type IRegisterResolver interface {
	Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (IServiceRegistry, error)
}

var registers = make(map[string]IRegisterResolver)

//Register 注册服务注册解析器
func Register(name string, resolver IRegisterResolver) {
	if resolver == nil {
		panic("registry: Register adapter is nil")
	}
	if _, ok := registers[name]; ok {
		panic("registry: Register called twice for adapter " + name)
	}
	registers[name] = resolver
}

//NewRegister 根据适配器名称及参数返回配置处理器
func NewRegister(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (IServiceRegistry, error) {
	resolver, ok := registers[adapter]
	if !ok {
		return nil, fmt.Errorf("registry: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, serverName, log, servers, cross)
}
