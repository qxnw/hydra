package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/disk"
)

//CollectDiskUP 收集硬盘使用率
func CollectDiskUP(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		diskInfo := disk.GetInfo()
		ip := net.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateDiskStatus(c, ctx, diskInfo.UsedPercent, "server", ip)
		response.SetContent(0, err)
		return
	}
}
