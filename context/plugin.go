package context

import (
	"time"

	"fmt"

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

type HandlerFunc func(name string, mode string, service string, c *Context) (Response, error)

func (h HandlerFunc) Handle(name string, mode string, service string, c *Context) (Response, error) {
	return h(name, mode, service, c)
}

type SHandlerFunc func(name string, mode string, service string, c *Context) (*StandardReponse, error)

func (h SHandlerFunc) Handle(name string, mode string, service string, c *Context) (*StandardReponse, error) {
	return h(name, mode, service, c)
}

type WHandlerFunc func(name string, mode string, service string, c *Context) (*WebReponse, error)

func (h WHandlerFunc) Handle(name string, mode string, service string, c *Context) (*WebReponse, error) {
	return h(name, mode, service, c)
}

//Handler context handler
type Handler interface {
	Handle(name string, mode string, service string, c *Context) (Response, error)
	Close() error
}
type MapHandler interface {
	Handle(name string, mode string, service string, c *Context) (*MapResponse, error)
	Close() error
}
type StandardHandler interface {
	Handle(name string, mode string, service string, c *Context) (*StandardReponse, error)
	Close() error
}
type ObjectHandler interface {
	Handle(name string, mode string, service string, c *Context) (*ObjectReponse, error)
	Close() error
}
type WebHandler interface {
	Handle(name string, mode string, service string, c *Context) (*WebReponse, error)
	Close() error
}

type Registry struct {
	Handlers map[string]interface{}
	Services []string
}

//NewRegistry 构建插件的注册中心
func NewRegistry() *Registry {
	r := &Registry{}
	r.Handlers = make(map[string]interface{})
	r.Services = make([]string, 0, 16)
	return r
}

//Register 注册处理程序
func (r *Registry) Register(name string, h interface{}) {
	for _, v := range r.Services {
		if v == name {
			panic(fmt.Sprintf("多次注册服务:%s", name))
		}
	}
	switch handler := h.(type) {
	case MapHandler, StandardHandler, WebHandler, ObjectHandler, Handler:
		r.Handlers[name] = handler
		r.Services = append(r.Services, name)
	default:
		panic(fmt.Sprintf("服务必须为Handler,MapHandler,StandardHandler,ObjectHandler,WebHandler:%s", name))
	}
}

func (r *Registry) Handle(name string, mode string, service string, c *Context) (Response, error) {
	response := GetStandardResponse()
	response.SetStatus(404)
	h, ok := r.Handlers[service]
	if !ok {
		return response, fmt.Errorf("未找到:%s", service)
	}
	switch handler := h.(type) {
	case MapHandler:
		return handler.Handle(name, mode, service, c)
	case StandardHandler:
		return handler.Handle(name, mode, service, c)
	case WebHandler:
		return handler.Handle(name, mode, service, c)
	case ObjectHandler:
		return handler.Handle(name, mode, service, c)
	case Handler:
		return handler.Handle(name, mode, service, c)
	default:
		return response, fmt.Errorf("未找到:%s", service)
	}
}
