package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/memory"
)

//CollectMemUP 收集内存使用率
func CollectMemUP(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		memoryInfo := memory.GetInfo()
		ip := net.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateMemStatus(c, ctx, memoryInfo.UsedPercent, "server", ip)
		response.SetContent(0, err)
		return
	}
}