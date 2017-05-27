package influx

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
)

type influxProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	serviceHandlers map[string]func(*context.Context) (string, error)
	dbs             cmap.ConcurrentMap
}

func newInfluxProxy() *influxProxy {
	r := &influxProxy{
		services: make([]string, 0, 4),
		dbs:      cmap.New(),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, error))
	r.serviceHandlers["/influx/save"] = r.save
	r.serviceHandlers["/influx/query"] = r.query
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *influxProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *influxProxy) Close() error {
	s.dbs.RemoveIterCb(func(key string, value interface{}) bool {
		client := value.(*influxdb.InfluxClient)
		client.Close()
		return true
	})
	return nil
}

//Handle
//save:从input中获取参数:measurement,tags,fields
//get:从input中获取参数:q
//从args中获取db参数
//influx配置：{"host":"http://192.168.0.185:8086","dataBase":"hydra","userName":"hydra","password":"123456"}

func (s *influxProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}

	content, err := s.serviceHandlers[service](ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: 200, Content: content}, nil

}

func (s *influxProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("不存在服务:%s", shortName)
}

type influxProxyResolver struct {
}

func (s *influxProxyResolver) Resolve() engine.IWorker {
	return newInfluxProxy()
}

func init() {
	engine.Register("influx", &influxProxyResolver{})
}
