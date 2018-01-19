package engines

import (
	"github.com/qxnw/hydra/engines/alarm"
	"github.com/qxnw/hydra/engines/file"
	"github.com/qxnw/hydra/engines/log"
	"github.com/qxnw/hydra/engines/mock"
	"github.com/qxnw/hydra/engines/rpc"
)

func init() {
	Component.AddMicroService("/file/upload", file.FileUpload(), "file")
	Component.AddMicroService("/log/info", log.WriteInfoLog(), "log")
	Component.AddMicroService("/log/error", log.WriteErrorLog(), "log")
	Component.AddMicroService("/mock/raw/request", mock.RawRequest(), "mock")
	Component.AddMicroService("/rpc/proxy", rpc.Proxy(), "rpc")
}
func (r *GroupEngine) loadServices() {
	r.AddMicroService("/alarm/collect/net/conn", alarm.NetConnNumCollect(r), "alarm")

	r.AddMicroService("/alarm/collect/api/server/response", alarm.HydraServerResponseCodeCollect(r, "api_server_reponse"), "alarm")
	r.AddMicroService("/alarm/collect/web/server/response", alarm.HydraServerResponseCodeCollect(r, "web_server_reponse"), "alarm")
	r.AddMicroService("/alarm/collect/rpc/server/response", alarm.HydraServerResponseCodeCollect(r, "rpc_server_reponse"), "alarm")
	r.AddMicroService("/alarm/collect/cron/server/response", alarm.HydraServerResponseCodeCollect(r, "cron_server_reponse"), "alarm")
	r.AddMicroService("/alarm/collect/mq/server/response", alarm.HydraServerResponseCodeCollect(r, "mq_consumer_reponse"), "alarm")

	r.AddMicroService("/alarm/collect/api/server/qps", alarm.HydraServerQPSCollect(r, "api_server_qps"), "alarm")
	r.AddMicroService("/alarm/collect/web/server/qps", alarm.HydraServerQPSCollect(r, "web_server_qps"), "alarm")
	r.AddMicroService("/alarm/collect/rpc/server/qps", alarm.HydraServerQPSCollect(r, "rpc_server_qps"), "alarm")
	r.AddMicroService("/alarm/collect/cron/server/qps", alarm.HydraServerQPSCollect(r, "cron_server_qps"), "alarm")
	r.AddMicroService("/alarm/collect/mq/server/qps", alarm.HydraServerQPSCollect(r, "mq_consumer_qps"), "alarm")

	r.AddMicroService("/alarm/collect/http/status", alarm.HTTPStatusCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/tcp/status", alarm.TCPStatusCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/sql/query", alarm.DBValueCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/registry/count", alarm.RegistryNodeCountCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/cpu/used", alarm.CPUUPCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/mem/used", alarm.MemUPCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/disk/used", alarm.DiskUPCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/net/conn", alarm.NetConnNumCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/nginx/error", alarm.NginxErrorCountCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/nginx/access", alarm.NginxAccessCountCollect(r), "alarm")
	r.AddMicroService("/alarm/collect/queue/count", alarm.QueueMessageCountCollect(r), "alarm")
	r.AddMicroService("/alarm/notify/send", alarm.SendAlarmNotify(r), "alarm")

}
