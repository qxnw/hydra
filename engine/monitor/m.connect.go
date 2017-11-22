package monitor

import (
	"fmt"
	"os/exec"
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
	err = updateNetConnectCountStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), int64(ncc), "server", ip, "max", fmt.Sprintf("%d", max))
	response.SetError(0, err)
	return
}
func (s *monitorProxy) getNetConnectCount() (v int, err error) {
	cmd1 := exec.Command("netstat", "-an")
	cmd2 := exec.Command("grep", "tcp")
	cmd3 := exec.Command("wc", "-l")
	cmds := []*exec.Cmd{cmd1, cmd2, cmd3}
	count, err := pipes.Run(cmds)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return v, nil
}
func (s *monitorProxy) getMaxOpenFiles() (v int, err error) {
	cmd1 := exec.Command("ulimit ", "-n")
	cmds := []*exec.Cmd{cmd1}
	count, err := pipes.Run(cmds)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return v, nil
}
