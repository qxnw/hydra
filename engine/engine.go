package engine

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
)

//IWorker 插件
type IWorker interface {
	Start(domain string, serverName string, serverType string) ([]string, error)
	Close() error
	context.EngineHandler
}

//IEngine 执行引擎
type IEngine interface {
	Start(domain string, serverName string, serverType string) ([]string, error)
	Handle(name string, method string, service string, c *context.Context) (*context.Response, error)
	Register(name string, p IWorker)
}

type IWorkerResolver interface {
	Resolve() IWorker
}

type standardEngine struct {
	plugins map[string]IWorker
	service map[string]IWorker
}

//NewStandardEngine 创建标准执行引擎
func NewStandardEngine() IEngine {
	e := &standardEngine{
		plugins: make(map[string]IWorker),
		service: make(map[string]IWorker),
	}
	for k, v := range resolvers {
		e.plugins[k] = v.Resolve()
	}
	return e
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
func (e *standardEngine) Register(name string, p IWorker) {
	if _, ok := e.plugins[name]; ok {
		panic("engine: Register called twice for adapter " + name)
	}
	e.plugins[name] = p
}

var resolvers map[string]IWorkerResolver

func init() {
	resolvers = make(map[string]IWorkerResolver)
}

//Register 注册插件
func Register(name string, p IWorkerResolver) {
	if _, ok := resolvers[name]; ok {
		panic("engine: Register called twice for adapter " + name)
	}
	resolvers[name] = p
}
