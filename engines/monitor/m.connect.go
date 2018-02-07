package monitor

import (
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

//CollectNetConnNum 收集网络连接数
func CollectNetConnNum(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		ip := net.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		ncc, err := getNetConnNum()
		if err != nil {
			return
		}
		max, err := getMaxOpenFiles()
		if err != nil {
			return
		}
		err = updateNetConnectCountStatus(c, ctx, int64(ncc), "server", ip, "max", fmt.Sprintf("%d", max))
		response.SetContent(0, err)
		return
	}
}
func getNetConnNum() (v int, err error) {
	count, err := pipes.BashRun(`netstat -an|grep tcp|wc -l`)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
func getMaxOpenFiles() (v int, err error) {
	count, err := pipes.BashRun("ulimit -n")
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
