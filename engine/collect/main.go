package collect

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

type collectProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	registryAddrs   string
	rpc             *rpc.Invoker
	registry        registry.Registry
	queryMap        map[string]string
	reportMap       map[string]string
	srvQueryMap     map[string]string
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
	r.srvQueryMap = make(map[string]string)
	r.collector = make(map[string]func(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (string, error))
	r.init()
	r.serviceHandlers = make(map[string]func(*context.Context) (string, int, error), 8)
	r.serviceHandlers["/collect/api/server/response"] = r.responseCollect("api_server_reponse")
	r.serviceHandlers["/collect/web/server/response"] = r.responseCollect("web_server_reponse")
	r.serviceHandlers["/collect/rpc/server/response"] = r.responseCollect("rpc_server_reponse")
	r.serviceHandlers["/collect/cron/server/response"] = r.responseCollect("cron_server_reponse")
	r.serviceHandlers["/collect/mq/consumer/response"] = r.responseCollect("mq_consumer_reponse")

	r.serviceHandlers["/collect/api/server/qps"] = r.requestQPSCollect("api_server_qps")
	r.serviceHandlers["/collect/web/server/qps"] = r.requestQPSCollect("web_server_qps")
	r.serviceHandlers["/collect/rpc/server/qps"] = r.requestQPSCollect("rpc_server_qps")
	r.serviceHandlers["/collect/cron/server/qps"] = r.requestQPSCollect("cron_server_qps")
	r.serviceHandlers["/collect/mq/consumer/qps"] = r.requestQPSCollect("mq_consumer_qps")

	r.serviceHandlers["/collect/http/status"] = r.httpCollect
	r.serviceHandlers["/collect/tcp/status"] = r.tcpCollect
	r.serviceHandlers["/collect/sql/query"] = r.dbCollect
	r.serviceHandlers["/collect/registry/count"] = r.registryCollect
	r.serviceHandlers["/collect/cpu/used"] = r.cpuCollect
	r.serviceHandlers["/collect/mem/used"] = r.memCollect
	r.serviceHandlers["/collect/disk/used"] = r.diskCollect
	r.serviceHandlers["/notify/send"] = r.notifySend

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
		return &context.Response{Status: types.DecodeInt(st, 0, 500)}, fmt.Errorf("engine:collect %s,%v", service, err)
	}
	return &context.Response{Status: types.DecodeInt(st, 0, 200), Content: content}, nil
}
func (s *collectProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:collect不存在服务:%s", shortName)
}

func (s *collectProxy) checkAndSave(ctx *context.Context, mode string, tf *transform.Transform, t int) (status int, err error) {
	status = 204
	db, err := s.getInfluxClient(ctx, "influxdb")
	if err != nil {
		return
	}

	query := tf.Translate(s.queryMap[mode])
	value, err := db.QueryMaps(query)
	if err != nil {
		return
	}
	if t == 0 {
		//上次无消息，则不上报
		if len(value) == 0 || len(value[0]) == 0 {
			return
		}
		//上次消息是成功不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "0" {
			return
		}
		//其它情况，上次消息是失败则上报
	} else {
		//上次消息是失败，但记录时间小于5分钟，则不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "1" {
			//fmt.Println("time:", value[0][0]["time"], value)
			lastTime, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", fmt.Sprintf("%v", value[0][0]["time"]))
			if err != nil {
				return 204, err
			}
			if time.Now().Sub(lastTime).Minutes() < 5 {
				return 204, nil
			}
		}
	}
	err = db.SendLineProto(tf.TranslateAll(s.reportMap[mode], true))
	if err != nil {
		return 500, err
	}
	return 200, nil
}

type collectProxyResolver struct {
}

func (s *collectProxyResolver) Resolve() engine.IWorker {
	return newCollectProxy()
}

func init() {
	engine.Register("collect", &collectProxyResolver{})
}
