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
	GetRegistry() registry.Registry
	GetServices() []string
	Fallback(name string, engine string, service string, c *context.Context) (rs context.Response, err error)
	Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)
	Close() error
}

//ServiceEngine 服务引擎
type ServiceEngine struct {
	*component.StandardComponent
	domain       string
	serverName   string
	serverType   string
	registryAddr string
	engines      []string
	*rpc.Invoker
	logger   *logger.Logger
	registry registry.Registry
	*varParamWatcher
	component.IComponentCache
	component.IComponentConf
	component.IComponentDB
	component.IComponentInfluxDB
	component.IComponentQueue
}

//NewServiceEngine 构建服务引擎
func NewServiceEngine(domain string, serverName string, serverType string, registryAddr string, logger *logger.Logger, engines ...string) (e *ServiceEngine, err error) {
	e = &ServiceEngine{domain: domain, serverName: serverName, serverType: serverType, registryAddr: registryAddr, logger: logger, engines: engines}
	e.StandardComponent = component.NewStandardComponent("sys.engine", e)
	e.Invoker = rpc.NewInvoker(domain, serverName, registryAddr)
	e.IComponentCache = component.NewStandardCache(e, "cache")
	e.IComponentConf = component.NewStandardConf(e)
	e.IComponentDB = component.NewStandardDB(e, "db")
	e.IComponentInfluxDB = component.NewStandardInfluxDB(e, "influx")
	e.IComponentQueue = component.NewStandardQueue(e, "queue")
	if e.registry, err = registry.NewRegistryWithAddress(registryAddr, logger); err != nil {
		return
	}
	e.varParamWatcher = newVarParamWatcher(domain, e.registry)
	e.loadEngineServices()
	if err = e.LoadComponents(fmt.Sprintf("./%s.so", domain),
		fmt.Sprintf("./%s.so", serverName),
		fmt.Sprintf("./%s_%s.so", domain, serverName)); err != nil {
		return
	}
	e.StandardComponent.AddRPCProxy(e.RPCProxy())
	err = e.StandardComponent.LoadServices()
	return
}

//GetServices 获取组件提供的所有服务
func (r *ServiceEngine) GetServices() []string {
	return r.GetGroupServices(component.GetGroupName(r.serverType))
}

//Execute 执行外部请求
func (r *ServiceEngine) Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
	service = formatName(service)
	if ctx.Request.CircuitBreaker.IsOpen() { //熔断开关打开，则自动降级
		response := context.GetStandardResponse()
		rf, err := r.StandardComponent.Fallback(name, engine, service, ctx)
		if rf != nil {
			if err == nil {
				return rf, nil
			}
			if err != component.ErrNotFoundService {
				return rf, err
			}
		}
		response.SetContent(ctx.Request.CircuitBreaker.GetDefStatus(), ctx.Request.CircuitBreaker.GetDefContent())
		return response, err
	}
	if rh, err := r.Handling(name, engine, service, ctx); err != nil {
		return rh, err
	}
	rx, err := r.Handle(name, engine, service, ctx)
	if err != nil {
		return rx, err
	}
	if rd, err := r.Handled(name, engine, service, ctx); err != nil {
		return rd, err
	}
	return rx, nil
}

//Handling 每次handle执行前执行
func (r *ServiceEngine) Handling(name string, engine string, service string, c *context.Context) (rs context.Response, err error) {
	c.SetRPC(r.Invoker)
	switch engine {
	case "rpc":
		return nil, nil
	case "", "*":
		if r.IsCustomerService(component.GetGroupName(r.serverType), service) {
			return nil, nil
		}
	default:
		for _, e := range r.engines {
			if e == engine && r.CheckTag(service, engine) && r.IsCustomerService(component.GetGroupName(r.serverType), service) {
				return nil, nil
			}
		}
	}
	response := context.GetStandardResponse()
	response.SetStatus(404)
	return response, fmt.Errorf("%s未找到服务:%s %v", r.Name, service, r.getServiceMeta(engine, service))
}
func (r *ServiceEngine) getServiceMeta(engine string, service string) string {
	return fmt.Sprintf(`engine:[%s]-%v
		group:[%s]-%s`, engine, r.GetTags(service), component.GetGroupName(r.serverType),
		r.GetGroups(service))
}

//GetRegistry 获取注册中心
func (r *ServiceEngine) GetRegistry() registry.Registry {
	return r.registry
}

//GetDomainName 获取域信息
func (r *ServiceEngine) GetDomainName() string {
	return r.domain
}

//GetServerName 获取服务器名称
func (r *ServiceEngine) GetServerName() string {
	return r.serverName
}

//GetServerType 获取服务器类型
func (r *ServiceEngine) GetServerType() string {
	return r.serverType
}

//Close 关闭引擎
func (r *ServiceEngine) Close() error {
	r.Invoker.Close()
	r.IComponentCache.Close()
	r.IComponentConf.Close()
	r.IComponentDB.Close()
	r.IComponentInfluxDB.Close()
	r.IComponentQueue.Close()
	r.varParamWatcher.Close()
	return nil
}
func formatName(name string) string {
	text := "/" + strings.Trim(strings.Trim(name, " "), "/")
	return strings.ToLower(text)
	//index := strings.LastIndex(text, "#")
	//if index < 0 {
	//return strings.ToLower(text)
	//}
	//return strings.ToLower(text[0:index])
}
