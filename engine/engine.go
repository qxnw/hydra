package engine

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
)

var (
	METHOD_NAME = []string{"request", "query", "delete", "update", "insert", "get", "post", "put", "delete", "main"}
)

//IWorker 插件
type IWorker interface {
	Has(service string) error
	Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) ([]string, error)
	Close() error
	context.EngineHandler
}

//IEngine 执行引擎
type IEngine interface {
	Start(domain string, serverName string, serverType string, rpcRegistryAddress string) ([]string, error)
	Handle(name string, mode string, service string, c *context.Context) (*context.Response, error)
	Register(name string, p IWorker)
	Close() error
}

type IWorkerResolver interface {
	Resolve() IWorker
}

type standardEngine struct {
	plugins    map[string]IWorker
	domain     string
	serverName string
}

//NewStandardEngine 创建标准执行引擎
func NewStandardEngine() IEngine {
	e := &standardEngine{
		plugins: make(map[string]IWorker),
	}
	for k, v := range resolvers {
		e.plugins[k] = v.Resolve()
	}
	return e
}

//启动引擎
func (e *standardEngine) Start(domain string, serverName string, serverType string, rpcRegistryAddrss string) (services []string, err error) {
	services = make([]string, 0, 8)
	e.domain = domain
	e.serverName = serverName
	invoker := rpc.NewRPCInvoker(domain, serverName, rpcRegistryAddrss)
	for _, p := range e.plugins {
		srvs, err := p.Start(domain, serverName, serverType, invoker)
		if err != nil {
			return nil, err
		}
		services = append(services, srvs...)
	}
	return services, nil
}
func (e *standardEngine) Close() error {
	for _, p := range e.plugins {
		p.Close()
	}
	return nil
}

//处理引擎
func (e *standardEngine) Handle(name string, mode string, service string, c *context.Context) (*context.Response, error) {
	fmt.Println("---------engine.handle:", mode, service, e.domain, e.serverName)
	svName := "/" + strings.Trim(strings.ToUpper(service), "/")
	if mode != "*" {
		worker, ok := e.plugins[mode]
		if !ok {
			return &context.Response{Status: 404}, fmt.Errorf("engine:未找到执行引擎:%s", mode)
		}
		err := worker.Has(svName)
		if err != nil {
			return &context.Response{Status: 404}, fmt.Errorf("engine:在引擎%s中未找到服务:%s(err:%v)", mode, svName, err)
		}
		return worker.Handle(name, mode, svName, c)
	}
	for d, worker := range e.plugins {
		if d == "rpc" {
			continue
		}
		err := worker.Has(svName)
		if err != nil {
			continue
		}

		return worker.Handle(name, mode, svName, c)
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
