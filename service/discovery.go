package service

import (
	"fmt"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//IDiscovery 服务发现组件
type IDiscovery interface {
	Notify() (chan []*registry.ServiceUpdater, error)
	ConsumeCurrent(string, string, string) error
	Consume(string, string, string, string, string) error
	Close() error
}

//IDiscoveryResolver 定义配置文件转换方法
type IDiscoveryResolver interface {
	Resolve(adapter string, domain string, tag string, log *logger.Logger, servers []string) (IDiscovery, error)
}

var discoveryResolvers = make(map[string]IDiscoveryResolver)

//Discovery 注册服务发现解析器
func Discovery(name string, resolver IDiscoveryResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := discoveryResolvers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	discoveryResolvers[name] = resolver
}

//NewDiscovery 根据适配器名称及参数返回配置处理器
func NewDiscovery(adapter string, domain string, tag string, log *logger.Logger, servers []string) (IDiscovery, error) {
	resolver, ok := discoveryResolvers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, tag, log, servers)
}
