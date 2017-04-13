package service

import "fmt"

//IRegister 服务注册组件
type IRegister interface {
	Register(serviceName string, endPointName string, data string) (string, error)
	Unregister(path string) error
	Close() error
}

//IRegisterResolver 定义配置文件转换方法
type IRegisterResolver interface {
	Resolve(adapter string, domain string, tag string, args ...string) (IRegister, error)
}

var registers = make(map[string]IRegisterResolver)

//Register 注册服务注册解析器
func Register(name string, resolver IRegisterResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := registers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	registers[name] = resolver
}

//NewRegister 根据适配器名称及参数返回配置处理器
func NewRegister(adapter string, domain string, system string, args ...string) (IRegister, error) {
	resolver, ok := registers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, system, args...)
}
