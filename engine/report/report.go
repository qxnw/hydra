package report

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/types"
)

type collectProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	registryAddrs   string
	registry        registry.Registry
	serviceHandlers map[string]func(*context.Context) (string, int, error)
}

func newCollectProxy() *collectProxy {
	r := &collectProxy{
		services: make([]string, 0, 8),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, int, error), 1)
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *collectProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	services = s.services
	s.registryAddrs = ctx.Registry
	s.registry, err = registry.NewRegistryWithAddress(ctx.Registry, ctx.Logger)
	return

}
func (s *collectProxy) Close() error {
	return nil
}
func (s *collectProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	content, st, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:collect %v", err)
		return &context.Response{Status: types.DecodeInt(st, 0, 500)}, err
	}
	return &context.Response{Status: types.DecodeInt(st, 0, 200), Content: content}, nil
}
func (s *collectProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:collect不存在服务:%s", shortName)
}

type collectProxyResolver struct {
}

func (s *collectProxyResolver) Resolve() engine.IWorker {
	return newCollectProxy()
}

func init() {
	engine.Register("collect", &collectProxyResolver{})
}
