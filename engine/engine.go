package engine

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
)

//IPlugin 插件
type IPlugin interface {
	Start(domain string, serverName string, serverType string) ([]string, error)
	Close() error
	context.EngineHandler
}

//IEngine 执行引擎
type IEngine interface {
	Start(domain string, serverName string, serverType string) ([]string, error)
	Handle(name string, method string, service string, c *context.Context) (*context.Response, error)
	Register(name string, p IPlugin)
}

type standardEngine struct {
	plugins map[string]IPlugin
	service map[string]IPlugin
}

//NewStandardEngine 创建标准执行引擎
func NewStandardEngine() IEngine {
	return &standardEngine{
		plugins: make(map[string]IPlugin),
		service: make(map[string]IPlugin),
	}
}

//启动引擎
func (e *standardEngine) Start(domain string, serverName string, serverType string) (services []string, err error) {
	for _, p := range e.plugins {
		services, err = p.Start(domain, serverName, serverType)
		if err != nil {
			return nil, err
		}
		for _, s := range services {
			e.service[s] = p
		}
	}
	return nil, nil
}
func (e *standardEngine) Close() error {
	for _, p := range e.plugins {
		p.Close()
	}
	return nil
}

//处理引擎
func (e *standardEngine) Handle(name string, method string, service string, c *context.Context) (*context.Response, error) {
	cmd := strings.ToUpper(fmt.Sprintf("%s.%s", service, method))
	svs, ok := e.service[cmd]
	if !ok {
		return nil, fmt.Errorf("engine:未找到执行引擎:%s", cmd)
	}
	return svs.Handle(cmd, method, service, c)
}

//Register 注册插件
func (e *standardEngine) Register(name string, p IPlugin) {
	if _, ok := e.plugins[name]; ok {
		panic("engine: Register called twice for adapter " + name)
	}
	e.plugins[name] = p
}

var DefaultEngine = NewStandardEngine()

//Start 启动引擎
func Start(domain string, serverName string, serverType string) ([]string, error) {
	return DefaultEngine.Start(domain, serverName, serverType)
}

//Handle 处理引擎
func Handle(name string, method string, service string, c *context.Context) (*context.Response, error) {
	return DefaultEngine.Handle(name, method, service, c)
}

//Register 注册插件
func Register(name string, p IPlugin) {
	DefaultEngine.Register(name, p)
}
