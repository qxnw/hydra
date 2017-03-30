package service

import "fmt"

type ServiceWatcher interface {
	Start() error
	Notify() chan []*ServiceUpdater
	Close() error
}

type ServiceUpdater struct {
	Value string
	Op    int
}

//ConfResolver 定义配置文件转换方法
type ConfResolver interface {
	Resolve(adapter string, domain string, tag string, args ...string) (ServiceWatcher, error)
}

var confResolvers = make(map[string]ConfResolver)

//Register 注册配置文件适配器
func Register(name string, resolver ConfResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := confResolvers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	confResolvers[name] = resolver
}

//NewWatcher 根据适配器名称及参数返回配置处理器
func NewWatcher(adapter string, domain string, tag string, args ...string) (ServiceWatcher, error) {
	resolver, ok := confResolvers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, tag, args...)
}
