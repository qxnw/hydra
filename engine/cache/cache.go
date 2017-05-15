package cache

import (
	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type cacheProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	serviceHandlers map[string]func(*context.Context) (string, error)
	dbs             cmap.ConcurrentMap
}

func newCacheProxy() *cacheProxy {
	r := &cacheProxy{
		dbs:      cmap.New(),
		services: make([]string, 0, 4),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, error))
	r.serviceHandlers["/cache/save"] = r.save
	r.serviceHandlers["/cache/get"] = r.get
	r.serviceHandlers["/cache/del"] = r.del
	r.serviceHandlers["/cache/delay"] = r.delay
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *cacheProxy) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	return s.services, nil
}
func (s *cacheProxy) Close() error {
	return nil
}
func (s *cacheProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}

	content, err := s.serviceHandlers[service](ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: 200, Content: content}, nil
}

func (s *cacheProxy) Has(shortName, fullName string) (err error) {
	return nil
}

type memcacheProxyResolver struct {
}

func (s *memcacheProxyResolver) Resolve() engine.IWorker {
	return newCacheProxy()
}

func init() {
	engine.Register("cache", &memcacheProxyResolver{})
}
