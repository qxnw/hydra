package context

import (
	"sync"

	"github.com/qxnw/lib4go/logger"
)

//Context 引擎执行上下文
type Context struct {
	Server  *server
	Request *Request
	RPC     *ContextRPC
	Log     logger.ILogger
}

//GetContext 从缓存池中获取一个context
func GetContext(queryString IData, form IData, param IData, setting IData, ext map[string]interface{}, logger *logger.Logger) *Context {
	c := contextPool.Get().(*Context)
	c.Request.reset(queryString, form, param, setting, ext)
	c.Log = logger
	return c
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
				RPC:     &ContextRPC{},
				Request: newRequest(),
			}
		},
	}
}

//Close 回收context
func (c *Context) Close() {
	c.Request.clear()
	contextPool.Put(c)
}
