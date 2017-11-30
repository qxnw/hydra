package monitor

import (
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
)

func (s *monitorProxy) cpuCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	cpuInfo := cpu.GetInfo()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateCPUStatus(ctx,cpuInfo.UsedPercent, "server", ip)
	response.SetError(0, err)
	return
}
