package servers

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

var IsDebug = false
var (
	ST_RUNNING   = "running"
	ST_STOP      = "stop"
	SRV_TP_API   = "api"
	SRV_FILE_API = "file"
	SRV_TP_RPC   = "rpc"
	SRV_TP_CRON  = "cron"
	SRV_TP_MQ    = "mq"
	SRV_TP_WEB   = "web"
)

//IRegistryServer 基于注册中心的服务器
type IRegistryServer interface {
	Notify(conf.Conf) error
	GetAddress() string
	GetServices() []string
	GetStatus() string
	Start() error
	Shutdown()
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
