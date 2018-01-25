package monitor

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/net"
)

//CollectNetPackages 收集网络数据包收发情况
func CollectNetPackages(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		netInfo, err := net.GetInfo()
		if err != nil {
			err = fmt.Errorf("获取网卡信息失败:%v", err)
			return
		}
		ip := xnet.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		for _, ni := range netInfo {
			err = updateNetRecvStatus(ctx, ni.BytesRecv, "server", ip, "name", ni.Name)
			if err != nil {
				response.SetError(0, err)
				return
			}
			err = updateNetSentStatus(ctx, ni.BytesSent, "server", ip, "name", ni.Name)
			if err != nil {
				response.SetError(0, err)
				return
			}
		}
		response.Success()
		return
	}
}
