package context

import (
	"sync"

	"fmt"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/transform"
)

type VarHandle func(tp string, name string) (string, error)

//Context 引擎执行上下文
type Context struct {
	Domain     string
	ServerName string
	ServerType string
	Input      *Input
	rpc        RPCInvoker
	RPC        *ContextRPC
	HTTP       *ContextHTTP
	logger.ILogger
}

//GetContext 从缓存池中获取一个context
func GetContext() *Context {
	return contextPool.Get().(*Context)
}

//SetInput 设置输入参数
func (c *Context) SetInput(input transform.ITransformGetter, param transform.ITransformGetter, body string, args map[string]string, ext map[string]interface{}) {
	c.Input = &Input{Input: input, Params: param, Body: body, Args: args, Ext: ext}
	c.ILogger, _ = c.getLogger()
	c.HTTP.Reset(c)
}

//SetRPC 根据输入的context创建插件的上下文对象
func (c *Context) SetRPC(rpc RPCInvoker) {
	c.rpc = rpc
	c.RPC.Reset(c)
}

func (c *Context) getLogger() (*logger.Logger, error) {
	if session, ok := c.Input.Ext["hydra_sid"]; ok {
		return logger.GetSession("hydra", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", c.Input.Ext)
}

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{
				Input: &Input{},
				RPC:   &ContextRPC{},
				HTTP:  &ContextHTTP{},
			}
		},
	}
}

//Close 回收context
func (c *Context) Close() {
	c.Input.Args = nil
	c.Input.Body = ""
	c.Input.Ext = nil
	c.Input.Input = nil
	c.Input.Params = nil
	c.Input = &Input{}
	contextPool.Put(c)
}
