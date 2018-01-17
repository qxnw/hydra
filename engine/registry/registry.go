package registry

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"
)

type registryProxy struct {
	ctx             *engine.EngineContext
	services        []string
	registry        registry.Registry
	serviceHandlers map[string]context.SHandlerFunc
}

func newRegistryProxy() *registryProxy {
	r := &registryProxy{
		services: make([]string, 0, 8),
	}
	r.serviceHandlers = make(map[string]context.SHandlerFunc, 8)
	r.serviceHandlers["/registry/save/all"] = r.saveAll
	r.serviceHandlers["/registry/get/value"] = r.getValue
	r.serviceHandlers["/registry/get/children"] = r.getChildren
	r.serviceHandlers["/registry/create/path"] = r.createPath
	r.serviceHandlers["/registry/create/ephemeral/path"] = r.createTempPath
	r.serviceHandlers["/registry/create/sequence/path"] = r.createSEQPath
	r.serviceHandlers["/registry/update/value"] = r.updateValue
	r.serviceHandlers["/registry/domain/copy"] = r.domainCopy
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *registryProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	s.registry, err = registry.NewRegistryWithAddress(ctx.Registry, ctx.Logger)
	return

}
func (s *registryProxy) Close() error {
	return nil
}

func (s *registryProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:registry %v", err)
		return
	}
	return
}
func (s *registryProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:registry不存在服务:%s", shortName)
}

type registryProxyResolver struct {
}

func (s *registryProxyResolver) Resolve() engine.IWorker {
	return newRegistryProxy()
}

func init() {
	engine.Register("registry", &registryProxyResolver{})
}
