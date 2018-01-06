package context

import (
	"sync"

	"fmt"

	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/transform"
)

type VarHandle func(tp string, name string) (string, error)

//Context 引擎执行上下文
type Context struct {
	Input    *Input
	rpc      RPCInvoker
	DB       *ContextDB
	Cache    *ContextCache
	MQ       *ContextMQ
	RPC      *ContextRPC
	HTTP     *ContextHTTP
	Influxdb *ContextInfluxdb
	logger.ILogger
}

//GetContext 从缓存池中获取一个context
func GetContext() *Context {
	return contextPool.Get().(*Context)
}

//SetInput 设置输入参数
func (c *Context) SetInput(input transform.ITransformGetter, param transform.ITransformGetter, body string, args map[string]string, ext map[string]interface{}) {
	c.Input = &Input{Input: input, Params: param, Body: body, Args: args, Ext: ext}
	if _, ok := c.Input.Ext["__test__"]; ok {
		c.ILogger = &tLogger{}
	}
	c.ILogger, _ = c.getLogger()
	c.DB.Reset(c)
	c.MQ.Reset(c)
	c.HTTP.Reset(c)
	c.Cache.Reset(c)
	c.Influxdb.Reset(c)
}

//SetRPC 根据输入的context创建插件的上下文对象
func (c *Context) SetRPC(rpc RPCInvoker) {
	c.rpc = rpc
	c.RPC.Reset(c)
}

//GetCache 获取缓存操作对象
func (c *Context) GetCache(names ...string) (cache.ICache, error) {
	return c.Cache.GetCache(names...)
}

//GetDB 获取数据库操作实例
func (c *Context) GetDB(names ...string) (*db.DB, error) {
	return c.DB.GetDB(names...)
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
				Input:    &Input{},
				DB:       &ContextDB{},
				Cache:    &ContextCache{},
				MQ:       &ContextMQ{},
				RPC:      &ContextRPC{},
				HTTP:     &ContextHTTP{},
				Influxdb: &ContextInfluxdb{},
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
