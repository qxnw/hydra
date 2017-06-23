package context

import (
	"encoding/json"
	"sync"

	"github.com/qxnw/lib4go/transform"
)

//EngineHandler context handle
type EngineHandler interface {
	Handle(name string, mode string, service string, c *Context) (*Response, error)
}

//Context 服务输出及Task执行的上下文
type Context struct {
	input InputArgs
	ext   map[string]interface{}
}

func (c *Context) GetInput() transform.ITransformGetter {
	return c.input.input
}
func (c *Context) GetArgs() map[string]string {
	return c.input.args
}
func (c *Context) GetBody() string {
	return c.input.body
}
func (c *Context) GetParams() transform.ITransformGetter {
	return c.input.params
}
func (c *Context) GetJson() string {
	return c.input.ToJson()
}
func (c *Context) GetExt() map[string]interface{} {
	return c.ext
}

//Response 响应
type Response struct {
	Content interface{}
	Status  int
	Params  map[string]interface{}
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
func (c *Context) Close() {
	c.input = InputArgs{}
	c.ext = make(map[string]interface{})
	contextPool.Put(c)
}

func GetContext() *Context {
	return contextPool.Get().(*Context)
}
func (c *Context) Set(input transform.ITransformGetter, param transform.ITransformGetter, body string, args map[string]string, ext map[string]interface{}) {
	c.input.input = input
	c.input.params = param
	c.input.args = args
	c.input.body = body
	c.ext = ext
}

//InputArgs 上下文输入参数
type InputArgs struct {
	input  transform.ITransformGetter `json:"input"`
	body   string                     `json:"body"`
	params transform.ITransformGetter `json:"params"`
	args   map[string]string          `json:"args"`
}

func (c *InputArgs) ToJson() string {
	data, _ := json.Marshal(c)
	return string(data)
}
