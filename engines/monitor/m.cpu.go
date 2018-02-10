package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
)

//CollectCPUUP 收集CPU使用率
func CollectCPUUP(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		cpuInfo := cpu.GetInfo()
		ip := net.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateCPUStatus(c, ctx, cpuInfo.UsedPercent, "server", ip)
		response.SetContent(0, err)
		return
	}
}
