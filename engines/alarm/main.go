package alarm

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

//LoadService 加载报警服务
func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddAutoflowService("/alarm/collect/api/server/response", HydraServerResponseCodeCollect(i, "api_server_reponse"), "alarm")
	r.AddAutoflowService("/alarm/collect/web/server/response", HydraServerResponseCodeCollect(i, "web_server_reponse"), "alarm")
	r.AddAutoflowService("/alarm/collect/rpc/server/response", HydraServerResponseCodeCollect(i, "rpc_server_reponse"), "alarm")
	r.AddAutoflowService("/alarm/collect/cron/server/response", HydraServerResponseCodeCollect(i, "cron_server_reponse"), "alarm")
	r.AddAutoflowService("/alarm/collect/mq/server/response", HydraServerResponseCodeCollect(i, "mq_consumer_reponse"), "alarm")

	r.AddAutoflowService("/alarm/collect/api/server/qps", HydraServerQPSCollect(i, "api_server_qps"), "alarm")
	r.AddAutoflowService("/alarm/collect/web/server/qps", HydraServerQPSCollect(i, "web_server_qps"), "alarm")
	r.AddAutoflowService("/alarm/collect/rpc/server/qps", HydraServerQPSCollect(i, "rpc_server_qps"), "alarm")
	r.AddAutoflowService("/alarm/collect/cron/server/qps", HydraServerQPSCollect(i, "cron_server_qps"), "alarm")
	r.AddAutoflowService("/alarm/collect/mq/server/qps", HydraServerQPSCollect(i, "mq_consumer_qps"), "alarm")

	r.AddAutoflowService("/alarm/collect/http/status", HTTPStatusCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/tcp/status", TCPStatusCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/sql/query", DBValueCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/registry/count", RegistryNodeCountCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/cpu/used", CPUUPCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/mem/used", MemUPCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/disk/used", DiskUPCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/net/conn", NetConnNumCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/nginx/error", NginxErrorCountCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/nginx/access", NginxAccessCountCollect(i), "alarm")
	r.AddAutoflowService("/alarm/collect/queue/count", QueueMessageCountCollect(i), "alarm")
	r.AddAutoflowService("/alarm/notify/send", SendAlarmNotify(i), "alarm")
}
func init() {
	engines.AddServiceLoader("alarm", LoadService)
}
