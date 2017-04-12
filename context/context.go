package context

import (
	"encoding/json"
	"sync"
)

//EngineHandler context handle
type EngineHandler interface {
	Handle(name string, method string, service string, c *Context) (*Response, error)
}

//Context 服务输出及Task执行的上下文
type Context struct {
	Input InputArgs
	Ext   map[string]interface{}
}

//Response 响应
type Response struct {
	Content string
	Status  int
	Params  map[string]interface{}
}

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &Context{Input: InputArgs{},
				Ext: make(map[string]interface{}),
			}
		},
	}
}
func (c *Context) Close() {
	c.Input = InputArgs{}
	c.Ext = make(map[string]interface{})
	contextPool.Put(c)
}

func GetContext() *Context {
	return contextPool.Get().(*Context)
}

//InputArgs 上下文输入参数
type InputArgs struct {
	Input  interface{} `json:"input"`
	Body   interface{} `json:"body"`
	Params interface{} `json:"params"`
	Args   interface{} `json:"args"`
}

func (c *InputArgs) ToJson() string {
	data, _ := json.Marshal(c)
	return string(data)
}
