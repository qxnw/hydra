package component

import (
	"fmt"

	"github.com/qxnw/hydra/context"
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

//GroupComponent 标准组件
type GroupComponent struct {
	Name            string                                            //组件名称
	funcs           map[string]map[string]func() (interface{}, error) //每个分组对应的服务及处理程序
	Handlers        map[string]interface{}                            //每个服务对应的处理程序
	Services        []string                                          //所有服务
	GroupServices   map[string][]string                               //每个分组包含的服务
	ServiceGroup    map[string][]string                               //每个服务对应的分组
	ServicePages    map[string][]string                               //每个服务对应的页面
	ServiceTagPages map[string]map[string][]string                    //服务对应的tag,及每个tag对应的页面
	ServicesTags    map[string][]string                               //服务对应的tag列表
	TagServices     map[string][]string                               //tag对应的服务列表
}

//NewGroupComponent 构建标准组件
func NewGroupComponent(componentName string) *GroupComponent {
	r := &GroupComponent{Name: componentName}
	r.funcs = make(map[string]map[string]func() (interface{}, error))
	r.Handlers = make(map[string]interface{})
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
func (r *GroupComponent) AddMicroService(service string, h interface{}, tags ...string) {
	r.addService(MicroService, service, h)
	r.addTags(service, tags...)
}

//AddAutoflowService 添加自动流程服务(供cron，mq方式调用)
func (r *GroupComponent) AddAutoflowService(service string, h interface{}, tags ...string) {
	r.addService(AutoflowService, service, h)
	r.addTags(service, tags...)
}

//AddPageService 添加页面服务
func (r *GroupComponent) AddPageService(service string, h interface{}, pages ...string) {
	r.addService(PageService, service, h)
	r.ServicePages[service] = pages
}

//AddTagPageService 添加带有标签的页面服务
func (r *GroupComponent) AddTagPageService(service string, h interface{}, tag string, pages ...string) {
	r.addService(PageService, service, h)
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
func (r *GroupComponent) AddCustomerService(service string, h interface{}, groupNames ...string) {
	if len(groupNames) == 0 {
		panic(fmt.Sprintf("服务:%s未指定分组名称", service))
	}
	for _, group := range groupNames {
		r.addService(group, service, h)
	}
}

//IsMicroService 是否是微服务
func (r *GroupComponent) IsMicroService(service string) bool {
	return r.IsCustomerService(MicroService, service)
}

//IsAutoflowService 是否是自动流程服务
func (r *GroupComponent) IsAutoflowService(service string) bool {
	return r.IsCustomerService(AutoflowService, service)
}

//IsPageService 是否是页面服务
func (r *GroupComponent) IsPageService(service string) bool {
	return r.IsCustomerService(PageService, service)
}

//IsCustomerService 是否是指定的分组服务
func (r *GroupComponent) IsCustomerService(group string, service string) bool {
	groups := r.GetGroups(service)
	for _, v := range groups {
		if v == group {
			return true
		}
	}
	return false
}
func (r *GroupComponent) addTags(service string, tags ...string) {
	r.ServicesTags[service] = append(r.ServicesTags[service], tags...)
	for _, tag := range tags {
		if _, ok := r.TagServices[tag]; !ok {
			r.TagServices[tag] = make([]string, 0, 2)
		}
		r.TagServices[tag] = append(r.TagServices[tag], service)
	}
}

//addService 添加服务处理程序
func (r *GroupComponent) addService(group string, service string, h interface{}) {
	if v, ok := h.(func() (interface{}, error)); ok {
		if _, ok := r.funcs[group]; !ok {
			r.funcs[group] = make(map[string]func() (interface{}, error))
		}
		if _, ok := r.funcs[group][service]; ok {
			panic(fmt.Sprintf("多次注册服务:%s", service))
		}
		r.funcs[group][service] = v
		return
	}
	r.register(group, service, h)
	return
}
func (r *GroupComponent) register(group string, name string, h interface{}) {
	for _, v := range r.GroupServices[group] {
		if v == name {
			panic(fmt.Sprintf("多次注册服务:%s", name))
		}
	}
	switch handler := h.(type) {
	case MapHandler, StandardHandler, WebHandler, ObjectHandler, Handler:
		if _, ok := r.Handlers[name]; !ok {
			r.Handlers[name] = handler
			r.Services = append(r.Services, name)
		}
		if _, ok := r.GroupServices[group]; !ok {
			r.GroupServices[group] = make([]string, 0, 2)
		}
		r.GroupServices[group] = append(r.GroupServices[group], name)

		if _, ok := r.ServiceGroup[name]; !ok {
			r.ServiceGroup[name] = make([]string, 0, 2)
		}
		r.ServiceGroup[name] = append(r.ServiceGroup[name], group)
	default:
		panic(fmt.Sprintf("服务必须为Handler,MapHandler,StandardHandler,ObjectHandler,WebHandler:%s", name))
	}
}

//LoadServices 加载所有服务
func (r *GroupComponent) LoadServices() error {
	for group, v := range r.funcs {
		for name, sv := range v {
			if h, ok := r.Handlers[name]; ok {
				r.register(group, name, h)
				continue
			}
			h, err := sv()
			if err != nil {
				return err
			}
			r.register(group, name, h)
		}
		delete(r.funcs, group)
	}
	return nil
}

//GetServices 获取组件提供的所有服务
func (r *GroupComponent) GetServices() []string {
	return r.Services
}

//GetGroupServices 根据分组获取服务
func (r *GroupComponent) GetGroupServices(group string) []string {
	return r.GroupServices[group]
}

//GetTagServices 根据tag获取服务列表
func (r *GroupComponent) GetTagServices(tag string) []string {
	return r.TagServices[tag]
}

//GetGroups 获取服务的分组列表
func (r *GroupComponent) GetGroups(service string) []string {
	return r.ServiceGroup[service]
}

//GetPages 获取服务的页面列表
func (r *GroupComponent) GetPages(service string) []string {
	return r.ServicePages[service]
}

//GetTagPages 获取服务的页面列表
func (r *GroupComponent) GetTagPages(service string, tagName string) []string {
	return r.ServiceTagPages[tagName][service]
}

//GetTags 获取服务的标签
func (r *GroupComponent) GetTags(service string) []string {
	return r.ServicesTags[service]
}

//CheckTag 检查服务标签是否匹配
func (r *GroupComponent) CheckTag(service string, tagName string) bool {
	for _, v := range r.ServicesTags[service] {
		if v == tagName {
			return true
		}
	}
	return false
}

//Handling 每次handle执行前执行
func (r *GroupComponent) Handling(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return nil, nil
}

//Handled 每次handle执行后执行
func (r *GroupComponent) Handled(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return nil, nil
}

//Handle 组件服务执行
func (r *GroupComponent) Handle(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
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
func (r *GroupComponent) Close() error {
	return nil
}

//GetGroupName 获取分组类型[api,rpc > micro mq,cron > autoflow, web > page,others > customer]
func GetGroupName(serverType string) string {
	switch serverType {
	case "api", "rpc":
		return MicroService
	case "mq", "cron":
		return AutoflowService
	case "web":
		return PageService
	}
	return CustomerService
}
