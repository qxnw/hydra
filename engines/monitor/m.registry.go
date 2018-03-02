package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
)

//CollectRegistryNodeCount 收集注册中心集点个数
func CollectRegistryNodeCount(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		path, err := ctx.Request.Setting.Get("path")
		if err != nil {
			return
		}
		data, _, err := c.GetRegistry().GetChildren(path)
		if err != nil {
			return
		}
		ip := xnet.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateRegistryStatus(c, ctx, int64(len(data)), "server", ip, "path", path)
		response.SetContent(0, err)
		return
	}
}
