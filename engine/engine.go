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
	Handle(name string, mode string, service string, c *context.Context) (*context.Response, error)
	Register(name string, p IWorker)
}

type IWorkerResolver interface {
	Resolve() IWorker
}

type standardEngine struct {
	plugins map[string]IWorker
	service map[string]map[string]IWorker
}

//NewStandardEngine 创建标准执行引擎
func NewStandardEngine() IEngine {
	e := &standardEngine{
		plugins: make(map[string]IWorker),
		service: make(map[string]map[string]IWorker),
	}
	for k, v := range resolvers {
		e.plugins[k] = v.Resolve()
	}
	return e
}

//启动引擎
func (e *standardEngine) Start(domain string, serverName string, serverType string) (services []string, err error) {
	for mode, p := range e.plugins {
		e.service[mode] = make(map[string]IWorker)
		services, err = p.Start(domain, serverName, serverType)
		if err != nil {
			return nil, err
		}
		for _, s := range services {
			e.service[mode][s] = p
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
func (e *standardEngine) Handle(name string, mode string, service string, c *context.Context) (*context.Response, error) {
	svName := strings.ToUpper(service)
	if mode != "*" {
		worker, ok := e.service[mode]
		if !ok {
			return &context.Response{Status: 404}, fmt.Errorf("engine:未找到执行引擎:%s", mode)
		}
		svs, ok := worker[svName]
		if !ok {
			return &context.Response{Status: 404}, fmt.Errorf("engine:在引擎%s未找到服务:%s", mode, svName)
		}
		return svs.Handle(name, mode, svName, c)
	}
	for _, worker := range e.service {
		svs, ok := worker[svName]
		if !ok {
			continue
		}
		return svs.Handle(name, mode, svName, c)
	}
	return &context.Response{Status: 404}, fmt.Errorf("engine:未找到服务:%s", svName)

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
