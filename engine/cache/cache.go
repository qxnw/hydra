package cache

import (
	"fmt"

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
	r.serviceHandlers["/cache/memcached/save"] = r.save
	r.serviceHandlers["/cache/memcached/get"] = r.get
	r.serviceHandlers["/cache/memcached/del"] = r.del
	r.serviceHandlers["/cache/memcached/delay"] = r.delay
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *cacheProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *cacheProxy) Close() error {
	return nil
}

//操作缓存
//从input参数中获取 key,value,expiresAt
//从args参数中获取 cache
//memcache.cache配置文件格式：{"server":"192.168.0.166:11212"}
func (s *cacheProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	content, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:cache.%v", err)
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: 200, Content: content}, nil
}

func (s *cacheProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:cache.不存在服务:%s", shortName)
}

type memcacheProxyResolver struct {
}

func (s *memcacheProxyResolver) Resolve() engine.IWorker {
	return newCacheProxy()
}

func init() {
	engine.Register("cache", &memcacheProxyResolver{})
}