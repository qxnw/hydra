package context

import (
	"sync"

	"fmt"

	"github.com/qxnw/lib4go/transform"
)

//Handler context handler
type Handler interface {
	Handle(name string, mode string, service string, c *Context) (*Response, error)
}

//Context 引擎执行上下文
type Context struct {
	input InputArgs
	ext   map[string]interface{}
}

//GetInput 获取输入参数
func (c *Context) GetInput() transform.ITransformGetter {
	return c.input.input
}

//GetArgs 获取配置参数
func (c *Context) GetArgs() map[string]string {
	return c.input.args
}

//GetBody 获取body参数
func (c *Context) GetBody(encoding ...string) (string, error) {
	if len(encoding) == 0 {
		return c.input.body, nil
	}
	if fun, ok := c.ext["__func_body_get_"].(func(ch string) (string, error)); ok {
		return fun(encoding[0])
	}
	return "", fmt.Errorf("无法根据%s格式转换数据:", encoding[0])
}

//GetParams 获取路由参数
func (c *Context) GetParams() transform.ITransformGetter {
	return c.input.params
}

//GetExt 获取扩展参数
func (c *Context) GetExt() map[string]interface{} {
	return c.ext
}

//Close 回收context
func (c *Context) Close() {
	c.input = InputArgs{}
	c.ext = make(map[string]interface{})
	contextPool.Put(c)
}

//GetContext 从缓存池中获取一个context
func GetContext() *Context {
	return contextPool.Get().(*Context)
}

//Set 设置输入参数
func (c *Context) Set(input transform.ITransformGetter, param transform.ITransformGetter, body string, args map[string]string, ext map[string]interface{}) {
	c.input.input = input
	c.input.params = param
	c.input.args = args
	c.input.body = body
	c.ext = ext
}

//InputArgs 上下文输入参数
type InputArgs struct {
	input  transform.ITransformGetter
	body   string
	params transform.ITransformGetter
	args   map[string]string
}

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{input: InputArgs{},
				ext: make(map[string]interface{}),
			}
		},
	}
}
