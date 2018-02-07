package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loadService(r *component.StandardComponent, i component.IContainer) {
	r.AddAutoflowService("/monitor/collect/cpu/used", CollectCPUUP(i), "monitor")
	r.AddAutoflowService("/monitor/collect/mem/used", CollectMemUP(i), "monitor")
	r.AddAutoflowService("/monitor/collect/disk/used", CollectDiskUP(i), "monitor")
	r.AddAutoflowService("/monitor/collect/net/status", CollectNetPackages(i), "monitor")
	r.AddAutoflowService("/monitor/collect/net/conn", CollectNetConnNum(i), "monitor")
	r.AddAutoflowService("/monitor/collect/http/status", CollectHTTPStatus(i), "monitor")
	r.AddAutoflowService("/monitor/collect/tcp/status", CollectTCPStatus(i), "monitor")
	r.AddAutoflowService("/monitor/collect/registry/count", CollectRegistryNodeCount(i), "monitor")
	r.AddAutoflowService("/monitor/collect/sql/query", CollectDBValue(i), "monitor")
	r.AddAutoflowService("/monitor/nginx/error/count", CollectNginxErrorNum(i), "monitor")
	r.AddAutoflowService("/monitor/nginx/access/count", CollectNginxAccessNum(i), "monitor")
	r.AddAutoflowService("/monitor/queue/count", CollectQueueMessageCount(i), "monitor")
}
func init() {
	engines.AddServiceLoader("monitor", loadService)
}
