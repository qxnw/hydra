package context

import (
	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/queue"
)

type IContainer interface {
	RPCInvoker

	conf.ISystemConf
	conf.IVarConf

	GetRegistry() registry.IRegistry
	GetCache(names ...string) (c cache.ICache, err error)
	GetDB(names ...string) (d db.IDB, err error)
	GetInflux(names ...string) (d influxdb.IInfluxClient, err error)
	GetQueue(names ...string) (q queue.IQueue, err error)
	Close() error
}

//Context 引擎执行上下文
type Context struct {
	Request   *Request
	Response  *Response
	RPC       *ContextRPC
	container IContainer
	Log       logger.ILogger
}

//GetContext 从缓存池中获取一个context
func GetContext(container IContainer, queryString IData, form IData, param IData, setting IData, ext map[string]interface{}, logger *logger.Logger) *Context {
	c := contextPool.Get().(*Context)
	c.Request.reset(queryString, form, param, setting, ext)
	c.Log = logger
	c.container = container
	return c
}

//GetContainer 获取当前容器
func (c *Context) GetContainer() IContainer {
	return c.container
}

//SetRPC 根据输入的context创建插件的上下文对象
func (c *Context) SetRPC(rpc RPCInvoker) {
	c.RPC.reset(c, rpc)
}

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{
				RPC:      &ContextRPC{},
				Request:  newRequest(),
				Response: NewResponse(),
			}
		},
	}
}

//Close 回收context
func (c *Context) Close() {
	c.Request.clear()
	c.Response.clear()
	c.container = nil
	contextPool.Put(c)
}
