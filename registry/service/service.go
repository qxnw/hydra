package service

import "fmt"

type ServiceWatcher interface {
	Notify() (chan []*ServiceUpdater, error)
	Publish(string, string, string) error
	ConsumeCurrent(string, string, string) error
	Consume(string, string, string, string, string) error
	Close() error
}

type ServiceUpdater struct {
	Value string
	Op    int
}

//ServiceResolver 定义配置文件转换方法
type ServiceResolver interface {
	Resolve(adapter string, domain string, tag string, args ...string) (ServiceWatcher, error)
}

var srvResolvers = make(map[string]ServiceResolver)

//Register 注册配置文件适配器
func Register(name string, resolver ServiceResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := srvResolvers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	srvResolvers[name] = resolver
}

//NewWatcher 根据适配器名称及参数返回配置处理器
func NewWatcher(adapter string, domain string, tag string, args ...string) (ServiceWatcher, error) {
	resolver, ok := srvResolvers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, tag, args...)
}
