package component

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

//StandardComponent 标准组件
type StandardComponent struct {
	Name     string
	funcs    map[string]func() (interface{}, error)
	Handlers map[string]interface{}
	Services []string
}

//NewStandardComponent 构建标准组件
func NewStandardComponent(componentName string) *StandardComponent {
	r := &StandardComponent{Name: componentName}
	r.funcs = make(map[string]func() (interface{}, error))
	r.Handlers = make(map[string]interface{})
	r.Services = make([]string, 0, 16)
	return r
}


//AddService 添加服务处理程序[服务可用于api,rpc]
func (r *StandardComponent) AddService(service string, h interface{}) {
	if v, ok := h.(func() (interface{}, error)); ok {
		if _, ok := r.funcs[service]; ok {
			panic(fmt.Sprintf("多次注册服务:%s", service))
		}
		r.funcs[service] = v
		return
	}
	r.register(service, h)
	return
}
func (r *StandardComponent) register(name string, h interface{}) {
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

//LoadServices 加载所有服务
func (r *StandardComponent) LoadServices() error {
	for name, v := range r.funcs {
		h, err := v()
		if err != nil {
			return err
		}
		r.register(name, h)
		delete(r.funcs, name)
	}
	return nil
}

//GetServices 获取组件提供的所有服务
func (r *StandardComponent) GetServices() []string {
	return r.Services
}

//Handling 每次handle执行前执行
func (r *StandardComponent) Handling(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return nil,nil
}

//Handled 每次handle执行后执行
func (r *StandardComponent) Handled(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return nil,nil
}

//Handle 组件服务执行
func (r *StandardComponent) Handle(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetStatus(404)
	h, ok := r.Handlers[service]
	if !ok {
		return response, fmt.Errorf("%s:未找到服务:%s", r.Name, service)
	}
	switch handler := h.(type) {
	case MapHandler:
		rs, err = handler.Handle(name, mode, service, c)
	case StandardHandler:
		rs, err = handler.Handle(name, mode, service, c)
	case WebHandler:
		rs, err = handler.Handle(name, mode, service, c)
	case ObjectHandler:
		rs, err = handler.Handle(name, mode, service, c)
	case Handler:
		rs, err = handler.Handle(name, mode, service, c)
	default:
		rs, err = response, fmt.Errorf("未找到服务:%s", service)
	}
	if err != nil {
		err = fmt.Errorf("%s:status:%d,err:%v", name, rs.GetStatus(err), err)
	}
	return
}

//Close 卸载组件
func (r *StandardComponent) Close() error {
	return nil
}
