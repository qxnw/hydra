package engines

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//IServiceEngine 服务引擎接口
type IServiceEngine interface {
	GetServices() []string
	Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)
	Close() error
}

var Component = component.NewGroupComponent("sys-engine")

type GroupEngine struct {
	component.GroupComponent
	domain       string
	serverName   string
	serverType   string
	registryAddr string
	engines      []string
	*rpc.Invoker
	logger   *logger.Logger
	registry registry.Registry
}

//NewGroupEngine 构建分组引擎
func NewGroupEngine(domain string, serverName string, serverType string, registryAddr string, logger *logger.Logger, engines ...string) (e *GroupEngine, err error) {
	e = &GroupEngine{domain: domain, serverName: serverName, serverType: serverType, registryAddr: registryAddr, logger: logger, engines: engines}
	e.GroupComponent = *Component
	e.Invoker = rpc.NewInvoker(domain, serverName, registryAddr)
	if e.registry, err = registry.NewRegistryWithAddress(registryAddr, logger); err != nil {
		return
	}
	e.loadServices()
	if err = e.LoadComponents(fmt.Sprintf("/%s.so", domain),
		fmt.Sprintf("/%s.so", serverName),
		fmt.Sprintf("/%s_%s.so", domain, serverName)); err != nil {
		return
	}
	err = e.GroupComponent.LoadServices()
	return
}

//GetServices 获取组件提供的所有服务
func (r *GroupEngine) GetServices() []string {
	return r.GetGroupServices(component.GetGroupName(r.serverType))
}

//Execute 执行外部请求
func (r *GroupEngine) Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
	service = formatName(service)
	if r, err := r.Handling(name, engine, service, ctx); err != nil {
		return r, err
	}
	rx, err := r.Handle(name, engine, service, ctx)
	if err != nil {
		return rx, err
	}
	if r, err := r.Handled(name, engine, service, ctx); err != nil {
		return r, err
	}
	return rx, nil
}

//Handling 每次handle执行前执行
func (r *GroupEngine) Handling(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	service = formatName(service)
	for _, e := range r.engines {
		if e == engine && r.CheckTag(service, e) && r.IsCustomerService(r.serverType, service) {
			return nil, nil
		}
	}
	response := context.GetStandardResponse()
	response.SetStatus(404)
	return response, fmt.Errorf("%s未找到服务：%v", r.Name, service)
}

//GetVarParam 获取配置参数
func (r *GroupEngine) GetVarParam(tp string, name string) (string, error) {
	buff, _, err := r.registry.GetValue(fmt.Sprintf("/%s/var/%s/%s", r.domain, tp, name))
	if err != nil {
		return "", err
	}
	return string(buff), nil
}

//GetRegistry 获取注册中心
func (r *GroupEngine) GetRegistry() registry.Registry {
	return r.registry
}

//GetDomainName 获取域信息
func (r *GroupEngine) GetDomainName() string {
	return r.domain
}

//GetServerName 获取服务器名称
func (r *GroupEngine) GetServerName() string {
	return r.serverName
}

//Close 关闭引擎
func (r *GroupEngine) Close() error {
	return nil
}
func formatName(name string) string {
	text := "/" + strings.Trim(name, "/")
	index := strings.LastIndex(text, "#")
	if index < 0 {
		return strings.ToLower(text)
	}
	return strings.ToLower(text[0:index])
}
