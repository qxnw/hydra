package engine

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/logger"
)

//IsDebug 当前服务器是处于调试模式
var IsDebug = false

var (
	METHOD_NAME = []string{"request", "query", "delete", "update", "insert", "create", "get", "post", "put", "delete", "main"}
	EXCLUDE     = []string{"lib", "conf", "config"}
)

//IWorker 插件
type IWorker interface {
	Has(shortName, fullName string) error
	Start(ctx *EngineContext) ([]string, error)
	Close() error
	context.EngineHandler
}

//IEngine 执行引擎
type IEngine interface {
	Start(domain string, serverName string, serverType string, rpcRegistryAddress string, logger *logger.Logger, extEngines ...string) ([]string, error)
	Handle(name string, mode string, service string, c *context.Context) (*context.Response, error)
	Register(name string, p IWorker)
	Close() error
}
type EngineContext struct {
	Domain     string
	ServerName string
	ServerType string
	Invoker    *rpc.RPCInvoker
	Registry   string
	Logger     *logger.Logger
}

type IWorkerResolver interface {
	Resolve() IWorker
}

type standardEngine struct {
	plugins    map[string]IWorker
	domain     string
	serverName string
	invoker    *rpc.RPCInvoker
	logger     *logger.Logger
}

//NewStandardEngine 创建标准执行引擎
func NewStandardEngine() IEngine {
	e := &standardEngine{
		plugins: make(map[string]IWorker),
	}
	return e
}

//启动引擎
func (e *standardEngine) Start(domain string, serverName string, serverType string, rpcRegistryAddrss string, logger *logger.Logger, extEngines ...string) (services []string, err error) {

	services = make([]string, 0, 8)
	e.domain = domain
	e.serverName = serverName
	e.logger = logger
	e.invoker = rpc.NewRPCInvoker(domain, serverName, rpcRegistryAddrss)
	//根据解析器生成引擎
	for k, v := range resolvers {
		hasExist := false
		for _, v := range extEngines {
			if strings.EqualFold(k, v) {
				hasExist = true
				break
			}
		}
		if !hasExist {
			continue
		}
		e.plugins[k] = v.Resolve()
	}
	//启动每个引擎
	engineContext := &EngineContext{Domain: domain,
		ServerName: serverName,
		ServerType: serverType,
		Invoker:    e.invoker,
		Registry:   rpcRegistryAddrss,
		Logger:     logger,
	}
	for _, p := range e.plugins {
		srvs, err := p.Start(engineContext)
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
	if e.invoker != nil {
		e.invoker.Close()
	}
	return nil
}

//处理引擎
func (e *standardEngine) Handle(name string, mode string, service string, c *context.Context) (*context.Response, error) {
	sName, fName := e.getServiceName(service)
	if mode != "*" {
		worker, ok := e.plugins[mode]
		if !ok {
			return &context.Response{Status: 404}, fmt.Errorf("engine:未找到执行引擎:%s", mode)
		}
		err := worker.Has(sName, fName)
		if err != nil {
			return &context.Response{Status: 404}, fmt.Errorf("engine:在引擎%s中未找到服务:%s(err:%v)", mode, fName, err)
		}
		return worker.Handle(sName, mode, fName, c)
	}
	for d, worker := range e.plugins {
		if d == "rpc" { //rpc为特殊服务，必须明确指定才能执行
			continue
		}
		err := worker.Has(sName, fName)
		if err != nil {
			continue
		}

		return worker.Handle(sName, mode, fName, c)
	}
	return &context.Response{Status: 404}, fmt.Errorf("engine:未找到服务:%s", sName)

}
func (e *standardEngine) getServiceName(name string) (sortName, fullName string) {
	text := "/" + strings.Trim(name, "/")
	index := strings.LastIndex(text, "#")
	if index < 0 {
		return strings.ToLower(text), strings.ToLower(text)
	}
	return strings.ToLower(text[0:index]), strings.ToLower(text[0:index]) + text[index:]
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
