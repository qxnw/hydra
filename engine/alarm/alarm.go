package alarm

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
)

type alarmProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	serviceHandlers map[string]func(*context.Context) (string, error)
	dbs             cmap.ConcurrentMap
}

func newAlarmProxy() *alarmProxy {
	r := &alarmProxy{
		services: make([]string, 0, 1),
		dbs:      cmap.New(),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, error))
	r.serviceHandlers["/alarm/influx/wx"] = r.influx2wx
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *alarmProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *alarmProxy) Close() error {
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

func (s *alarmProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	content, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:alarm %v", err)
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: 200, Content: content}, nil

}

func (s *alarmProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:alarm不存在服务:%s", shortName)
}

type alarmProxyResolver struct {
}

func (s *alarmProxyResolver) Resolve() engine.IWorker {
	return newAlarmProxy()
}

func init() {
	engine.Register("alarm", &alarmProxyResolver{})
}
