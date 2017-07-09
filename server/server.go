package server

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
)

//IsDebug 当前服务器是处于调试模式
var IsDebug = false
var (
	ST_RUNNING   = "running"
	ST_STOP      = "stop"
	SRV_TP_API   = "api"
	SRV_FILE_API = "file"
	SRV_TP_RPC   = "rpc"
	SRV_TP_CRON  = "cron"
	SRV_TP_MQ    = "mq"
)

type IServiceRegistry interface {
	RegisterTempNode(serviceName string, endPointName string, data string) (string, error)
	RegisterSeqNode(path string, data string) (string, error)
	Unregister(path string) error
	Close() error
}

//IHydraServer 服务器接口，可通过单个变量和配置文件方式设置服务器启用参数
type IHydraServer interface {
	Notify(conf.Conf) error
	GetAddress() string
	GetStatus() string
	Start() error
	Shutdown()
}

//IServerAdapter 服务解析器
type IServerAdapter interface {
	Resolve(c context.EngineHandler, r IServiceRegistry, conf conf.Conf) (IHydraServer, error)
}

var serverResolvers = make(map[string]IServerAdapter)

//Register 注册服务适配器
func Register(name string, resolver IServerAdapter) {
	if _, ok := serverResolvers[name]; ok {
		panic("server: Register called twice for adapter " + name)
	}
	serverResolvers[name] = resolver
}

//NewServer 根据适配器名称生成服务
func NewServer(adapter string, c context.EngineHandler, r IServiceRegistry, conf conf.Conf) (IHydraServer, error) {
	if resolver, ok := serverResolvers[adapter]; ok {
		return resolver.Resolve(c, r, conf)
	}
	return nil, fmt.Errorf("server: unknown adapter name %q (forgotten import?)", adapter)

}
