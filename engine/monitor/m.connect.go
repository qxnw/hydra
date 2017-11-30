package monitor

import (
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

func (s *monitorProxy) netConnectCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	ncc, err := s.getNetConnectCount()
	if err != nil {
		return
	}
	max, err := s.getMaxOpenFiles()
	if err != nil {
		return
	}
	err = updateNetConnectCountStatus(ctx, int64(ncc), "server", ip, "max", fmt.Sprintf("%d", max))
	response.SetError(0, err)
	return
}
func (s *monitorProxy) getNetConnectCount() (v int, err error) {
	count, err := pipes.BashRun(`netstat -an|grep tcp|wc -l`)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
func (s *monitorProxy) getMaxOpenFiles() (v int, err error) {
	count, err := pipes.BashRun("ulimit -n")
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
