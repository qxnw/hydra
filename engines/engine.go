package engines

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/rpc"
	"github.com/qxnw/lib4go/logger"
)

//IServiceEngine 服务引擎接口
type IServiceEngine interface {
	GetRegistry() registry.IRegistry
	GetServices() []string
	Fallback(name string, engine string, service string, c *context.Context) (rs interface{})
	Execute(name string, engine string, service string, ctx *context.Context) (rs interface{})
	Close() error
}

//ServiceEngine 服务引擎
type ServiceEngine struct {
	*component.StandardComponent
	cHandler component.IComponentHandler
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

	if err = e.loadEngineServices(); err != nil {
		return nil, err
	}
	if err = e.LoadComponents(fmt.Sprintf("./%s.so", conf.GetPlatName()),
		fmt.Sprintf("./%s.so", conf.GetSysName()),
		fmt.Sprintf("./%s_%s.so", conf.GetPlatName(), conf.GetSysName())); err != nil {
		return
	}
	e.StandardComponent.AddRPCProxy(e.RPCProxy())
	err = e.StandardComponent.LoadServices()
	return
}

//SetHandler 设置handler
func (r *ServiceEngine) SetHandler(h component.IComponentHandler) error {
	if h == nil {
		return nil
	}
	r.cHandler = h
	svs := h.GetServices()
	for group, handlers := range svs {
		for name, handler := range handlers {
			r.StandardComponent.AddCustomerService(name, handler, group)
		}
	}
	return r.StandardComponent.LoadServices()
}

//UpdateVarConf 更新var配置参数
func (r *ServiceEngine) UpdateVarConf(conf conf.IServerConf) {
	r.SetVarConf(conf.GetVarConfClone())
	r.SetSubConf(conf.GetSubConfClone())
}

//GetServices 获取组件提供的所有服务
func (r *ServiceEngine) GetServices() []string {
	return r.GetGroupServices(component.GetGroupName(r.GetServerType())...)
}

//Execute 执行外部请求
func (r *ServiceEngine) Execute(name string, engine string, service string, ctx *context.Context) (rs interface{}) {
	service = formatName(service)
	if ctx.Request.CircuitBreaker.IsOpen() { //熔断开关打开，则自动降级
		rf := r.StandardComponent.Fallback(name, engine, service, ctx)
		if r, ok := rf.(error); ok && r == component.ErrNotFoundService {
			ctx.Response.MustContent(ctx.Request.CircuitBreaker.GetDefStatus(), ctx.Request.CircuitBreaker.GetDefContent())
		}
		return rf
	}
	if rh := r.Handling(name, engine, service, ctx); ctx.Response.HasError(rh) {
		return rh
	}

	if r.cHandler != nil && r.cHandler.GetHandling() != nil {
		if rh := r.cHandler.GetHandling()(name, engine, service, ctx); ctx.Response.HasError(rh) {
			return rh
		}
	}

	if rs = r.Handle(name, engine, service, ctx); ctx.Response.HasError(rs) {
		return rs
	}
	if r.cHandler != nil && r.cHandler.GetHandled() != nil {
		if rh := r.cHandler.GetHandled()(name, engine, service, ctx); ctx.Response.HasError(rh) {
			return rh
		}
	}
	if rd := r.Handled(name, engine, service, ctx); ctx.Response.HasError(rd) {
		return rd
	}
	return rs
}

//Handling 每次handle执行前执行
func (r *ServiceEngine) Handling(name string, engine string, service string, c *context.Context) (rs interface{}) {
	c.SetRPC(r.Invoker)
	switch engine {
	case "rpc":
		return nil
	default:
		if r.IsCustomerService(service,component.GetGroupName(r.GetServerType())...) {
			return nil
		}
	}
	c.Response.SetStatus(404)
	return fmt.Errorf("%s未找到服务:%s", r.Name, service)
}

//GetRegistry 获取注册中心
func (r *ServiceEngine) GetRegistry() registry.IRegistry {
	return r.registry
}

//Close 关闭引擎
func (r *ServiceEngine) Close() error {
	r.StandardComponent.Close()
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
