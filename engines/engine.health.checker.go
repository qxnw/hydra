package engines

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/sysinfo/memory"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

func healthCheck(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetMapResponse()
		data := make(map[string]interface{})
		data["cpu_used_precent"] = fmt.Sprintf("%.2f", cpu.GetInfo(time.Millisecond*200).UsedPercent)
		data["mem_used_precent"] = fmt.Sprintf("%.2f", memory.GetInfo().UsedPercent)
		data["disk_used_precent"] = fmt.Sprintf("%.2f", disk.GetInfo().UsedPercent)
		data["app_memory"] = memory.GetAPPMemory()
		data["plat-name"] = c.GetPlatName()
		data["system-name"] = c.GetSysName()
		data["server-type"] = c.GetServerType()
		data["cluster-name"] = c.GetClusterName()
		data["net-conn-cnt"], _ = getNetConnectNum()
		response.SetContent(200, data)
		return
	}
}
func getNetConnectNum() (v int, err error) {
	count, err := pipes.BashRun(`netstat -an|grep tcp|wc -l`)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
func serverLoader() ServiceLoader {
	return func(component *component.StandardComponent, container component.IContainer) {
		component.AddMicroService("/_server/health/check", healthCheck)
	}
}
func init() {
	AddServiceLoader(serverLoader())
}