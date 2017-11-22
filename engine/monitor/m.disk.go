package monitor

import (
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/disk"
)

func (s *monitorProxy) diskCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	diskInfo := disk.GetInfo()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateDiskStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), diskInfo.UsedPercent, "server", ip)
	response.SetError(0, err)
	return
}
