package alarm

import (
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
	"github.com/qxnw/lib4go/utility"
)

type collectProxy struct {
	ctx             *engine.EngineContext
	services        []string
	registryAddrs   string
	registry        registry.Registry
	queryMap        map[string]string
	reportMap       map[string]string
	srvQueryMap     map[string]string
	reportSQL       string
	serviceHandlers map[string]context.SHandlerFunc
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
	r.serviceHandlers = make(map[string]context.SHandlerFunc, 8)
	r.serviceHandlers["/alarm/collect/api/server/response"] = r.responseCollect("api_server_reponse")
	r.serviceHandlers["/alarm/collect/web/server/response"] = r.responseCollect("web_server_reponse")
	r.serviceHandlers["/alarm/collect/rpc/server/response"] = r.responseCollect("rpc_server_reponse")
	r.serviceHandlers["/alarm/collect/cron/server/response"] = r.responseCollect("cron_server_reponse")
	r.serviceHandlers["/alarm/collect/mq/consumer/response"] = r.responseCollect("mq_consumer_reponse")

	r.serviceHandlers["/alarm/collect/api/server/qps"] = r.requestQPSCollect("api_server_qps")
	r.serviceHandlers["/alarm/collect/web/server/qps"] = r.requestQPSCollect("web_server_qps")
	r.serviceHandlers["/alarm/collect/rpc/server/qps"] = r.requestQPSCollect("rpc_server_qps")
	r.serviceHandlers["/alarm/collect/cron/server/qps"] = r.requestQPSCollect("cron_server_qps")
	r.serviceHandlers["/alarm/collect/mq/consumer/qps"] = r.requestQPSCollect("mq_consumer_qps")

	r.serviceHandlers["/alarm/collect/http/status"] = r.httpCollect
	r.serviceHandlers["/alarm/collect/tcp/status"] = r.tcpCollect
	r.serviceHandlers["/alarm/collect/sql/query"] = r.dbCollect
	r.serviceHandlers["/alarm/collect/registry/count"] = r.registryCollect
	r.serviceHandlers["/alarm/collect/cpu/used"] = r.cpuCollect
	r.serviceHandlers["/alarm/collect/mem/used"] = r.memCollect
	r.serviceHandlers["/alarm/collect/disk/used"] = r.diskCollect
	r.serviceHandlers["/alarm/collect/net/conn"] = r.netConnectCountCollect
	r.serviceHandlers["/alarm/collect/nginx/error"] = r.nginxErrorCountCollect
	r.serviceHandlers["/alarm/collect/nginx/access"] = r.nginxAccessCountCollect
	r.serviceHandlers["/alarm/collect/queue/count"] = r.queueCountCollect

	r.serviceHandlers["/alarm/notify/send"] = r.notifySend

	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *collectProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	services = s.services
	s.registryAddrs = ctx.Registry

	s.registry, err = registry.NewRegistryWithAddress(ctx.Registry, ctx.Logger)
	return

}
func (s *collectProxy) Close() error {
	return nil
}

func (s *collectProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:collect %s,%v", service, err)
		return
	}
	return
}
func (s *collectProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:collect不存在服务:%s", shortName)
}

func (s *collectProxy) checkAndSave(ctx *context.Context, mode string, tf *transform.Transform, t int) (status int, err error) {
	status = 204
	db, err := ctx.Influxdb.GetClient("influxdb")
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
			lastTime, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", fmt.Sprintf("%v", value[0][0]["time"]))
			if err != nil {
				return 204, err
			}
			if time.Now().Sub(lastTime).Minutes() < 5 {
				return 204, nil
			}
		}
	}
	sql := s.reportMap[mode]
	return s.save2Influxdb(sql, tf, db)
}
func (s *collectProxy) save2Influxdb(sql string, tf *transform.Transform, db *influxdb.InfluxClient) (int, error) {
	sqls := strings.Split(sql, " ")
	measurement := sqls[0]
	tagsMap, err := utility.GetMapWithQuery(strings.Replace(sqls[1], ",", "&", -1))
	if err != nil {
		return 500, err
	}
	filedsMap, err := utility.GetMapWithQuery(strings.Replace(sqls[2], ",", "&", -1))
	if err != nil {
		return 500, err
	}
	tags := make(map[string]string)
	for k, v := range tagsMap {
		tags[k] = tf.TranslateAll(v, true)
	}
	fileds := make(map[string]interface{})
	for k, v := range filedsMap {
		fileds[k] = tf.TranslateAll(v, true)
	}
	if len(tags) == 0 || len(fileds) == 0 {
		err = fmt.Errorf("tags 或 fileds的个数不能为0")
		return 500, err
	}
	err = db.Send(measurement, tags, fileds)
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
	engine.Register("alarm", &collectProxyResolver{})
}
