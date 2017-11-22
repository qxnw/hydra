package monitor

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/net"
)

func (s *monitorProxy) netCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	netInfo, err := net.GetInfo()
	if err != nil {
		err = fmt.Errorf("获取网卡信息失败:%v", err)
		return
	}
	ip := xnet.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	for _, ni := range netInfo {
		err = updateNetRecvStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), ni.BytesRecv, "server", ip, "name", ni.Name)
		if err != nil {
			response.SetError(0, err)
			return
		}
		err = updateNetSentStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), ni.BytesSent, "server", ip, "name", ni.Name)
		if err != nil {
			response.SetError(0, err)
			return
		}
	}
	response.Success()
	return
}
