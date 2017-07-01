package collect

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/types"
)

type collectProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	registryAddrs   string
	rpc             *rpc.RPCInvoker
	registry        registry.Registry
	queryMap        map[string]string
	reportMap       map[string]string
	reportSQL       string
	serviceHandlers map[string]func(*context.Context) (string, int, error)
	collector       map[string]func(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (string, error)
}

func newCollectProxy() *collectProxy {
	r := &collectProxy{
		services: make([]string, 0, 1),
	}
	r.queryMap = make(map[string]string)
	r.reportMap = make(map[string]string)
	r.collector = make(map[string]func(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (string, error))
	r.reportSQL = `select * from hydra_collector where "time">now() - @time order by time`
	r.queryMap["http"] = `select value from hydra_collector where "type"='http' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["tcp"] = `select value from hydra_collector where "type"='tcp' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["registry"] = `select value from hydra_collector where "type"='registry' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["db"] = `select value from hydra_collector where "type"='db' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["cpu"] = `select value from hydra_collector where "type"='cpu' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["mem"] = `select value from hydra_collector where "type"='mem' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["disk"] = `select value from hydra_collector where "type"='disk' and "host"='@host' and "time">'now()-6h' order by time desc limit 1`

	r.reportMap["http"] = "hydra_collector,type=http,host=@host,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["tcp"] = "hydra_collector,type=tcp,host=@host,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["registry"] = "hydra_collector,type=registry,host=@host,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["db"] = "hydra_collector,type=db,host=@host,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["cpu"] = "hydra_collector,type=cpu,host=@host,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["mem"] = "hydra_collector,type=mem,host=@host,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["disk"] = "hydra_collector,type=disk,host=@host,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	r.collector["http"] = r.httpCollect
	r.collector["tcp"] = r.tcpCollect
	r.collector["registry"] = r.registryCollect
	//r.collector["db"] = r.dbCollect
	r.collector["cpu"] = r.cpuCollect
	r.collector["mem"] = r.memCollect
	r.collector["disk"] = r.diskCollect
	r.serviceHandlers = make(map[string]func(*context.Context) (string, int, error), 8)
	r.serviceHandlers["/collect/http/status"] = r.httpHandle
	r.serviceHandlers["/collect/tcp/status"] = r.tcpHandle
	r.serviceHandlers["/collect/sql/query"] = r.dbHandle
	r.serviceHandlers["/collect/registry/count"] = r.registryHandle
	r.serviceHandlers["/collect/cpu/used"] = r.cpuHandle
	r.serviceHandlers["/collect/mem/used"] = r.memHandle
	r.serviceHandlers["/collect/disk/used"] = r.diskHandle
	r.serviceHandlers["/notify/send"] = r.notify
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
	s.rpc = ctx.Invoker
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
		err = fmt.Errorf("engine:collect %s,%v", service, err)
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
