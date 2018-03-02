package component

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/types"
)

const (
	//MicroService 微服务
	MicroService = "__micro_"
	//AutoflowService 自动流程服务
	AutoflowService = "__autoflow_"

	//PageService 页面服务
	PageService = "__page_"

	//CustomerService 自定义服务
	CustomerService = "__customer_"
)

//ErrNotFoundService 未找到服务
var ErrNotFoundService = errors.New("未找到服务")

//StandardComponent 标准组件
type StandardComponent struct {
	Container        IContainer
	Name             string                            //组件名称
	funcs            map[string]map[string]interface{} //每个分组对应的服务及处理程序
	Handlers         map[string]interface{}            //每个服务对应的处理程序
	FallbackHandlers map[string]interface{}            //每个服务对应的降级处理程序
	Services         []string                          //所有服务
	GroupServices    map[string][]string               //每个分组包含的服务
	ServiceGroup     map[string][]string               //每个服务对应的分组
	ServicePages     map[string][]string               //每个服务对应的页面
	ServiceTagPages  map[string]map[string][]string    //服务对应的tag,及每个tag对应的页面
	ServicesTags     map[string][]string               //服务对应的tag列表
	TagServices      map[string][]string               //tag对应的服务列表
}

//NewStandardComponent 构建标准组件
func NewStandardComponent(componentName string, c IContainer) *StandardComponent {
	r := &StandardComponent{Name: componentName, Container: c}
	r.funcs = make(map[string]map[string]interface{})
	r.Handlers = make(map[string]interface{})
	r.FallbackHandlers = make(map[string]interface{})
	r.GroupServices = make(map[string][]string)
	r.ServiceGroup = make(map[string][]string)
	r.Services = make([]string, 0, 2)
	r.ServicePages = make(map[string][]string)
	r.ServiceTagPages = make(map[string]map[string][]string)
	r.ServicesTags = make(map[string][]string)
	r.TagServices = make(map[string][]string)
	return r
}

//AddMicroService 添加微服务(供http,rpc方式调用)
func (r *StandardComponent) AddMicroService(service string, h interface{}, tags ...string) {
	r.addService(MicroService, service, h)
	r.addTags(service, tags...)
}

//AddAutoflowService 添加自动流程服务(供cron，mq方式调用)
func (r *StandardComponent) AddAutoflowService(service string, h interface{}, tags ...string) {
	r.addService(AutoflowService, service, h)
	r.addTags(service, tags...)
}

//AddPageService 添加页面服务
func (r *StandardComponent) AddPageService(service string, h interface{}, pages ...string) {
	r.addService(PageService, service, h)
	r.ServicePages[service] = pages
}

//AddRPCProxy 添加RPC代理
func (r *StandardComponent) AddRPCProxy(h interface{}) {
	r.addService(MicroService, "__rpc_", h)
	r.addService(AutoflowService, "__rpc_", h)
}

//AddTagPageService 添加带有标签的页面服务
func (r *StandardComponent) AddTagPageService(service string, h interface{}, tag string, pages ...string) {
	r.addService(PageService, service, h)
	r.ServicePages[service] = pages
	if _, ok := r.ServiceTagPages[tag]; !ok {
		r.ServiceTagPages[tag] = make(map[string][]string)
	}
	if _, ok := r.ServiceTagPages[tag][service]; !ok {
		r.ServiceTagPages[tag][service] = make([]string, 0, 2)
	}
	r.ServiceTagPages[tag][service] = pages
	r.addTags(service, tag)
}

//AddCustomerService 添加自定义分组服务
func (r *StandardComponent) AddCustomerService(service string, h interface{}, groupNames ...string) {
	if len(groupNames) == 0 {
		panic(fmt.Sprintf("服务:%s未指定分组名称", service))
	}
	for _, group := range groupNames {
		r.addService(group, service, h)
	}
}

//AddCustomerTagService 添加自定义分组服务
func (r *StandardComponent) AddCustomerTagService(service string, h interface{}, tag string, groupNames ...string) {
	r.AddCustomerService(service, h, groupNames...)
	r.addTags(service, tag)
}

//AddCustomerTagsService 添加自定义分组服务
func (r *StandardComponent) AddCustomerTagsService(service string, h interface{}, tags []string, groupNames ...string) {
	r.AddCustomerService(service, h, groupNames...)
	r.addTags(service, tags...)
}

//IsMicroService 是否是微服务
func (r *StandardComponent) IsMicroService(service string) bool {
	return r.IsCustomerService(MicroService, service)
}

//IsAutoflowService 是否是自动流程服务
func (r *StandardComponent) IsAutoflowService(service string) bool {
	return r.IsCustomerService(AutoflowService, service)
}

//IsPageService 是否是页面服务
func (r *StandardComponent) IsPageService(service string) bool {
	return r.IsCustomerService(PageService, service)
}

//IsCustomerService 是否是指定的分组服务
func (r *StandardComponent) IsCustomerService(group string, service string) bool {
	groups := r.GetGroups(service)
	for _, v := range groups {
		if v == group {
			return true
		}
	}
	return false
}
func (r *StandardComponent) addTags(service string, tags ...string) {
	r.ServicesTags[service] = append(r.ServicesTags[service], tags...)
	for _, tag := range tags {
		if _, ok := r.TagServices[tag]; !ok {
			r.TagServices[tag] = make([]string, 0, 2)
		}
		r.TagServices[tag] = append(r.TagServices[tag], service)
	}
}

//addService 添加服务处理程序
func (r *StandardComponent) addService(group string, service string, h interface{}) {
	r.register(group, service, h)
	return
}
func (r *StandardComponent) registerAddService(name string, group string, handler interface{}) {
	if _, ok := r.Handlers[name]; !ok {
		r.Handlers[name] = handler
		r.Services = append(r.Services, name)
	}
	if strings.HasPrefix(name, "__") {
		return
	}
	if _, ok := r.GroupServices[group]; !ok {
		r.GroupServices[group] = make([]string, 0, 2)
	}
	r.GroupServices[group] = append(r.GroupServices[group], name)

	if _, ok := r.ServiceGroup[name]; !ok {
		r.ServiceGroup[name] = make([]string, 0, 2)
	}
	r.ServiceGroup[name] = append(r.ServiceGroup[name], group)
}
func (r *StandardComponent) register(group string, name string, h interface{}) {
	for _, v := range r.GroupServices[group] {
		if v == name {
			panic(fmt.Sprintf("多次注册服务:%s:%v", name, r.GroupServices[group]))
		}
	}

	//注册get,post,put,delete,handle服务
	found := false
	switch handler := h.(type) {
	case GetHandler:
		var f ServiceFunc = handler.GetHandle
		r.registerAddService(filepath.Join(name, "get"), group, f)
		found = true
	}
	switch handler := h.(type) {
	case PostHandler:
		var f ServiceFunc = handler.PostHandle
		r.registerAddService(filepath.Join(name, "post"), group, f)
		found = true
	}
	switch handler := h.(type) {
	case PutHandler:
		var f ServiceFunc = handler.PutHandle
		r.registerAddService(filepath.Join(name, "put"), group, f)
		found = true
	}
	switch handler := h.(type) {
	case DeleteHandler:
		var f ServiceFunc = handler.DeleteHandle
		r.registerAddService(filepath.Join(name, "delete"), group, f)
		found = true
	}
	switch h.(type) {
	case Handler:
		r.registerAddService(name, group, h)
		found = true
	}
	if !found {
		r.checkFuncType(name, h)
		if _, ok := r.funcs[group]; !ok {
			r.funcs[group] = make(map[string]interface{})
		}
		if _, ok := r.funcs[group][name]; ok {
			panic(fmt.Sprintf("多次注册服务:%s", name))
		}
		r.funcs[group][name] = h
	}

	//处理降级服务

	//get降级服务
	switch handler := h.(type) {
	case GetFallbackHandler:
		name := filepath.Join(name, "get")
		var f FallbackServiceFunc = handler.GetFallback
		if _, ok := r.FallbackHandlers[name]; !ok {
			r.FallbackHandlers[name] = f
		}
	}

	//post降级服务
	switch handler := h.(type) {
	case PostFallbackHandler:
		name := filepath.Join(name, "post")
		var f FallbackServiceFunc = handler.PostFallback
		if _, ok := r.FallbackHandlers[name]; !ok {
			r.FallbackHandlers[name] = f
		}
	}

	//put降级服务
	switch handler := h.(type) {
	case PutFallbackHandler:
		name := filepath.Join(name, "put")
		var f FallbackServiceFunc = handler.PutFallback
		if _, ok := r.FallbackHandlers[name]; !ok {
			r.FallbackHandlers[name] = f
		}
	}

	//delete降级服务
	switch handler := h.(type) {
	case DeleteFallbackHandler:
		name := filepath.Join(name, "delete")
		var f FallbackServiceFunc = handler.DeleteFallback
		if _, ok := r.FallbackHandlers[name]; !ok {
			r.FallbackHandlers[name] = f
		}
	}

	//通用降级服务
	switch handler := h.(type) {
	case FallbackHandler:
		if _, ok := r.FallbackHandlers[name]; !ok {
			r.FallbackHandlers[name] = handler
		}
	}

}
func (r *StandardComponent) checkFuncType(name string, h interface{}) {
	fv := reflect.ValueOf(h)
	if fv.Kind() != reflect.Func {
		panic(fmt.Sprintf("服务:%s必须为Handler,MapHandler,StandardHandler,ObjectHandler,WebHandler, Handler, MapServiceFunc, StandardServiceFunc, WebServiceFunc, ServiceFunc:%v", name, h))
	}
	tp := reflect.TypeOf(h)
	if tp.NumIn() > 2 || tp.NumOut() == 0 || tp.NumOut() > 2 {
		panic(fmt.Sprintf("服务:%s只能包含最多1个输入参数，最多2个返回值", name))
	}
	if tp.NumIn() == 1 {
		if tp.In(0).Name() != "IContainer" {
			panic(fmt.Sprintf("服务:%s输入参数必须为component.IContainer类型(%s)", name, tp.In(0).Name()))
		}
	}
	if tp.NumOut() == 2 {
		if tp.Out(1).Name() != "error" {
			panic(fmt.Sprintf("服务:%s的2个返回值必须为error类型", name))
		}
	}
}
func (r *StandardComponent) callFuncType(name string, h interface{}) (i interface{}, err error) {
	fv := reflect.ValueOf(h)
	tp := reflect.TypeOf(h)
	var rvalue []reflect.Value
	if tp.NumIn() == 1 {
		ivalue := make([]reflect.Value, 0, 1)
		ivalue = append(ivalue, reflect.ValueOf(r.Container))
		rvalue = fv.Call(ivalue)
	} else {
		rvalue = fv.Call(nil)
	}
	if len(rvalue) == 0 || len(rvalue) > 2 {
		panic(fmt.Sprintf("%s类型错误,返回值只能有1个(handler)或2个（Handler,error）", name))
	}
	if len(rvalue) > 1 {
		if rvalue[1].Interface() != nil {
			if err, ok := rvalue[1].Interface().(error); ok {
				return nil, err
			}
		}
	}

	return rvalue[0].Interface(), nil
}

//LoadServices 加载所有服务
func (r *StandardComponent) LoadServices() error {
	for group, v := range r.funcs {
		for name, sv := range v {
			if h, ok := r.Handlers[name]; ok {
				r.register(group, name, h)
				continue
			}
			rt, err := r.callFuncType(name, sv)
			if err != nil {
				return err
			}
			r.register(group, name, rt)
		}
		delete(r.funcs, group)
	}
	return nil
}

//GetServices 获取组件提供的所有服务
func (r *StandardComponent) GetServices() []string {
	return r.Services
}

//GetGroupServices 根据分组获取服务
func (r *StandardComponent) GetGroupServices(group string) []string {
	return r.GroupServices[group]
}

//GetTagServices 根据tag获取服务列表
func (r *StandardComponent) GetTagServices(tag string) []string {
	return r.TagServices[tag]
}

//GetGroups 获取服务的分组列表
func (r *StandardComponent) GetGroups(service string) []string {
	return r.ServiceGroup[service]
}

//GetPages 获取服务的页面列表
func (r *StandardComponent) GetPages(service string) []string {
	return r.ServicePages[service]
}

//GetTagPages 获取服务的页面列表
func (r *StandardComponent) GetTagPages(service string, tagName string) []string {
	return r.ServiceTagPages[tagName][service]
}

//GetTags 获取服务的标签
func (r *StandardComponent) GetTags(service string) []string {
	return r.ServicesTags[service]
}

//GetFallbackHandlers 获取fallback处理程序
func (r *StandardComponent) GetFallbackHandlers() map[string]interface{} {
	return r.FallbackHandlers
}

//AddFallbackHandlers 添加降级函数
func (r *StandardComponent) AddFallbackHandlers(f map[string]interface{}) {
	for k, v := range f {
		r.FallbackHandlers[k] = v
	}
}

//CheckTag 检查服务标签是否匹配
func (r *StandardComponent) CheckTag(service string, tagName string) bool {
	for _, v := range r.ServicesTags[service] {
		if v == tagName {
			return true
		}
	}
	return false
}

//Handling 每次handle执行前执行
func (r *StandardComponent) Handling(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	return nil, nil
}

//Handled 每次handle执行后执行
func (r *StandardComponent) Handled(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	return nil, nil
}

//GetHandler 获取服务的处理函数
func (r *StandardComponent) GetHandler(engine string, service string, method string) (interface{}, bool) {
	switch engine {
	case "rpc":
		r, ok := r.Handlers["__rpc_"]
		return r, ok
	default:
		if r, ok := r.Handlers[filepath.Join(service, method)]; ok {
			return r, ok
		}
		r, ok := r.Handlers[service]
		return r, ok
	}
}

//Handle 组件服务执行
func (r *StandardComponent) Handle(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetStatus(404)
	method := c.Request.Ext.GetMethod()
	h, ok := r.GetHandler(engine, service, method)
	if !ok {
		return response, fmt.Errorf("%s:未找到服务:%s", r.Name, service)
	}
	switch handler := h.(type) {
	case Handler:
		rs, err = handler.Handle(name, engine, service, c)
	default:
		rs, err = response, fmt.Errorf("未找到服务:%s", service)
	}
	if err != nil {
		status := 500
		if rs != nil && !reflect.ValueOf(rs).IsNil() {
			status = rs.GetStatus()
		}
		err = fmt.Errorf("%s:status:%d,err:%v", r.Name, types.DecodeInt(status, 0, 500, status), err)
	}
	return
}

//GetFallbackHandler 获取失败降级处理函数
func (r *StandardComponent) GetFallbackHandler(engine string, service string, method string) (interface{}, bool) {
	if f, ok := r.FallbackHandlers[filepath.Join(service, method)]; ok {
		return f, ok
	}
	f, ok := r.FallbackHandlers[service]
	return f, ok

}

//Fallback 降级处理
func (r *StandardComponent) Fallback(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetStatus(404)
	h, ok := r.GetFallbackHandler(engine, service, c.Request.Ext.GetMethod())
	if !ok {
		return response, ErrNotFoundService
	}
	switch handler := h.(type) {
	case FallbackHandler:
		rs, err = handler.Fallback(name, engine, service, c)
	default:
		rs, err = response, fmt.Errorf("未找到服务:%s", service)
	}
	if err != nil {
		status := 500
		if rs != nil && !reflect.ValueOf(rs).IsNil() {
			status = rs.GetStatus()
		}
		err = fmt.Errorf("%s:status:%d,err:%v", r.Name, types.DecodeInt(status, 0, 500, status), err)
	}
	return
}

//Close 卸载组件
func (r *StandardComponent) Close() error {
	r.funcs = nil
	r.Handlers = nil
	r.GroupServices = nil
	r.ServiceGroup = nil
	r.Services = nil
	r.ServicePages = nil
	r.ServiceTagPages = nil
	r.ServicesTags = nil
	r.TagServices = nil
	return nil
}

//GetGroupName 获取分组类型[api,rpc > micro mq,cron > autoflow, web > page,others > customer]
func GetGroupName(serverType string) string {
	switch serverType {
	case "api", "rpc":
		return MicroService
	case "mqc", "cron":
		return AutoflowService
	case "web":
		return PageService
	}
	return CustomerService
}
