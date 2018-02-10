package servers

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

var IsDebug = false
var (
	ST_RUNNING   = "running"
	ST_STOP      = "stop"
	ST_PAUSE     = "pause"
	SRV_TP_API   = "api"
	SRV_FILE_API = "file"
	SRV_TP_RPC   = "rpc"
	SRV_TP_CRON  = "cron"
	SRV_TP_MQ    = "mq"
	SRV_TP_WEB   = "web"
)

type IServer interface {
	GetAddress() string
	GetServices() []string
	GetStatus() string
	Run(string) error
	SetJWT(enable bool, name string, mode string, secret string, exclude []string, expireAt int64)
	SetHost(host string)
	SetMetric(host string, dataBase string, userName string, password string, cron string)
	SetHeader(headers map[string]string)
	StopMetric()
	Shutdown(time.Duration)
}

//IRegistryServer 基于注册中心的服务器
type IRegistryServer interface {
	Notify(conf.Conf) error
	Start() error
	GetAddress() string
	GetServices() []string
	GetStatus() string
	Shutdown()
}

type IExecuter interface {
	Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)
}

type IExecuteHandler func(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)

func (i IExecuteHandler) Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
	return i(name, engine, service, ctx)
}

//IRegistryEngine 基于注册中心的执行引擎
type IRegistryEngine interface {
	GetRegistry() registry.Registry
	GetServices() []string
	Fallback(name string, engine string, service string, c *context.Context) (rs context.Response, err error)
	Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)
	Close() error
}

//IServerResolver 服务器生成器
type IServerResolver interface {
	Resolve(c IRegistryEngine, conf conf.Conf, log *logger.Logger) (IRegistryServer, error)
}
type IServerResolverHandler func(c IRegistryEngine, conf conf.Conf, log *logger.Logger) (IRegistryServer, error)

//Resolve 创建服务器实例
func (i IServerResolverHandler) Resolve(c IRegistryEngine, conf conf.Conf, log *logger.Logger) (IRegistryServer, error) {
	return i(c, conf, log)
}

var resolvers = make(map[string]IServerResolver)

//Register 注册服务器生成器
func Register(identifier string, resolver IServerResolver) {
	if _, ok := resolvers[identifier]; ok {
		panic("server: Register called twice for identifier: " + identifier)
	}
	resolvers[identifier] = resolver
}

//NewRegistryServer 根据服务标识创建服务器
func NewRegistryServer(identifier string, c IRegistryEngine, conf conf.Conf, log *logger.Logger) (IRegistryServer, error) {
	if resolver, ok := resolvers[identifier]; ok {
		return resolver.Resolve(c, conf, log)
	}
	return nil, fmt.Errorf("server: unknown identifier name %q (forgotten import?)", identifier)
}

//Trace 打印跟踪信息
func Trace(print func(f string, args ...interface{}), serverName string, args ...interface{}) {
	if !IsDebug {
		return
	}
	print("%s:%s", serverName, args)
}

//Tracef 根据格式打印跟踪信息
func Tracef(print func(f string, args ...interface{}), format string, args ...interface{}) {
	if !IsDebug {
		return
	}
	print(format, args...)
}

//TraceIf 根据条件打印跟踪信息
func TraceIf(b bool, okPrint func(f string, args ...interface{}), print func(f string, args ...interface{}), serverName string, args ...interface{}) {
	if !IsDebug {
		return
	}
	if b {
		okPrint("%s:%s", serverName, args)
		return
	}
	print("%s:%s", serverName, args)
}
