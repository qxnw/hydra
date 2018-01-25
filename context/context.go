package context

import (
	"sync"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/transform"
)

//Context 引擎执行上下文
type Context struct {
	Server  *server
	Request *Request
	RPC     *ContextRPC
	Log     logger.ILogger
}

//GetContext 从缓存池中获取一个context
func GetContext() *Context {
	return contextPool.Get().(*Context)
}

//SetInput 设置输入参数
func (c *Context) SetInput(input IData, param IData, setting map[string]string, ext map[string]interface{}) {
	c.Request.reset(input, param, transform.NewMap(setting).Data, ext)
	c.Log = logger.GetSession("hydra", c.Request.Ext.GetUUID())
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
