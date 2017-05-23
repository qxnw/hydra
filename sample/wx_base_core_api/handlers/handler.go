package handlers

import "github.com/qxnw/hydra_plugin/plugins"

//Handler 处理程序接口
type Handler interface {
	Handle(service string, c plugins.Context, invoker plugins.RPCInvoker) (status int, result string, err error)
}

//ServiceHandlers 服务处理程序列表
var ServiceHandlers map[string]Handler

//Services 当前提供的服务列表
var Services []string

func init() {
	ServiceHandlers = make(map[string]Handler)
	Services = make([]string, 0, 16)
	Register("/wx/notify", newWXNotify())

}

//Register 注册处理程序
func Register(name string, handler Handler) {
	if _, ok := ServiceHandlers[name]; ok {
		panic("wx_base_core_api: Register called twice for adapter " + name)
	}
	ServiceHandlers[name] = handler
	Services = append(Services, name)
}
