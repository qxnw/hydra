package monitor

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type monitorProxy struct {
	ctx             *engine.EngineContext
	services        []string
	registryAddrs   string
	serviceHandlers map[string]context.SHandlerFunc
}

func newMonitorProxy() *monitorProxy {
	r := &monitorProxy{
		services: make([]string, 0, 1),
	}
	r.serviceHandlers = make(map[string]context.SHandlerFunc)
	r.serviceHandlers["/monitor/collect/cpu/used"] = r.cpuCollect
	r.serviceHandlers["/monitor/collect/mem/used"] = r.cpuCollect
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *monitorProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	services = s.services
	s.registryAddrs = ctx.Registry
	return

}
func (s *monitorProxy) Close() error {
	return nil
}
func (s *monitorProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:monitor %s,%v", service, err)
		return
	}
	return
}
func (s *monitorProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:monitor不存在服务:%s", shortName)
}

type monitorResolver struct {
}

func (s *monitorResolver) Resolve() engine.IWorker {
	return newMonitorProxy()
}

func init() {
	engine.Register("monitor", &monitorResolver{})
}
