package engines

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//IServiceEngine 服务引擎接口
type IServiceEngine interface {
	GetRegistry() registry.IRegistry
	GetServices() []string
	Fallback(name string, engine string, service string, c *context.Context) (rs context.Response, err error)
	Execute(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error)
	Close() error
}

//ServiceEngine 服务引擎
type ServiceEngine struct {
	*component.StandardComponent
	conf.IServerConf
	registryAddr string
	*rpc.Invoker
	logger   *logger.Logger
	registry registry.IRegistry
	component.IComponentCache
	component.IComponentDB
	component.IComponentInfluxDB
	component.IComponentQueue
}

//NewServiceEngine 构建服务引擎
func NewServiceEngine(conf conf.IServerConf, registryAddr string, logger *logger.Logger) (e *ServiceEngine, err error) {
	e = &ServiceEngine{IServerConf: conf, registryAddr: registryAddr, logger: logger}
	e.StandardComponent = component.NewStandardComponent("sys.engine", e)
	e.Invoker = rpc.NewInvoker(conf.GetPlatName(), conf.GetSysName(), registryAddr)
	e.IComponentCache = component.NewStandardCache(e, "cache")
	e.IComponentDB = component.NewStandardDB(e, "db")
	e.IComponentInfluxDB = component.NewStandardInfluxDB(e, "influx")
	e.IComponentQueue = component.NewStandardQueue(e, "queue")
	if e.registry, err = registry.NewRegistryWithAddress(registryAddr, logger); err != nil {
		return
	}

	e.loadEngineServices()
	if err = e.LoadComponents(fmt.Sprintf("./%s.so", conf.GetPlatName()),
		fmt.Sprintf("./%s.so", conf.GetSysName()),
		fmt.Sprintf("./%s_%s.so", conf.GetPlatName(), conf.GetSysName())); err != nil {
		return
	}
	e.StandardComponent.AddRPCProxy(e.RPCProxy())
	err = e.StandardComponent.LoadServices()
	return
}

//UpdateVarConf 更新var配置参数
func (r *ServiceEngine) UpdateVarConf(conf conf.IServerConf) {
	r.SetVarConf(conf.GetVarConfClone())
	r.SetSubConf(conf.GetSubConfClone())
}

//GetServices 获取组件提供的所有服务
func (r *ServiceEngine) GetServices() []string {
	return r.GetGroupServices(component.GetGroupName(r.GetServerType()))
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
		if r.IsCustomerService(component.GetGroupName(r.GetServerType()), service) {
			return nil, nil
		}
	default:
		if r.IsCustomerService(component.GetGroupName(r.GetServerType()), service) {
			return nil, nil
		}

	}
	response := context.GetStandardResponse()
	response.SetStatus(404)
	return response, fmt.Errorf("%s未找到服务:%s %v", r.Name, service, r.getServiceMeta(engine, service))
}
func (r *ServiceEngine) getServiceMeta(engine string, service string) string {
	return fmt.Sprintf(`services：%v engine:[%s]-%v
		group:[%s]-%s`, r.GetServices(), engine, r.GetTags(service), component.GetGroupName(r.GetServerType()),
		r.GetGroups(service))
}

//GetRegistry 获取注册中心
func (r *ServiceEngine) GetRegistry() registry.IRegistry {
	return r.registry
}

//Close 关闭引擎
func (r *ServiceEngine) Close() error {
	r.Invoker.Close()
	r.IComponentCache.Close()
	r.IComponentDB.Close()
	r.IComponentInfluxDB.Close()
	r.IComponentQueue.Close()
	return nil
}
func formatName(name string) string {
	text := "/" + strings.Trim(strings.Trim(name, " "), "/")
	return strings.ToLower(text)
}
func appendEngines(engines []string, ext ...string) []string {
	addEngine := make([]string, 0, len(ext))
	for _, n := range ext {
		var b bool
		for _, en := range engines {
			if en == n {
				b = true
				continue
			}
		}
		if !b {
			addEngine = append(addEngine, n)
		}
	}
	return append(engines, addEngine...)
}
