package monitor

import (
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/memory"
)

func (s *monitorProxy) memCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	memoryInfo := memory.GetInfo()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateMemStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), ip, memoryInfo.UsedPercent)
	response.SetError(0, err)
	return
}
