package context

import (
	"time"

	"github.com/qxnw/lib4go/rpc"
	"github.com/qxnw/lib4go/transform"
)

type IContext interface {
	GetInput() transform.ITransformGetter
	GetArgs() map[string]string
	GetBody(encoding ...string) (string, error)
	GetParams() transform.ITransformGetter
	GetExt() map[string]interface{}
}

type RPCInvoker interface {
	PreInit(services ...string) (err error)
	RequestFailRetry(service string, input map[string]string, times int) (status int, result string, params map[string]string, err error)
	Request(service string, input map[string]string, failFast bool) (status int, result string, param map[string]string, err error)
	AsyncRequest(service string, input map[string]string, failFast bool) rpc.IRPCResponse
	WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error
}

type Worker interface {
	GetServices() []string
	Handler
}

type HandlerFunc func(name string, mode string, service string, c *Context) (*Response, error)

func (h HandlerFunc) Handle(name string, mode string, service string, c *Context) (*Response, error) {
	return h(name, mode, service, c)
}

//Handler context handler
type Handler interface {
	Handle(name string, mode string, service string, c *Context) (*Response, error)
	Close() error
}
type Registry struct {
	ServiceHandlers map[string]Handler
	Services        []string
}

//NewRegistry 构建插件的注册中心
func NewRegistry() *Registry {
	r := &Registry{}
	r.ServiceHandlers = make(map[string]Handler)
	r.Services = make([]string, 0, 16)
	return r
}

//Register 注册处理程序
func (r *Registry) Register(name string, handler Handler) {
	if _, ok := r.ServiceHandlers[name]; ok {
		panic("Register called twice for adapter " + name)
	}
	r.ServiceHandlers[name] = handler
	r.Services = append(r.Services, name)
}
