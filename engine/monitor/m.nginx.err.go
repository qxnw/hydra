package monitor

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

func (s *monitorProxy) nginxErrorCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	c, t, err := s.getNginxErrorCount()
	if err != nil {
		return
	}
	err = updateNginxErrorCount(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), int64(c), "server", ip, "minute", t)
	response.SetError(0, err)
	return
}
func (s *monitorProxy) getNginxErrorCount() (m int, tm string, err error) {
	tm = time.Now().Add(-1 * time.Minute).Format("15:04")
	cmd1 := exec.Command("cat", "/usr/local/nginx/logs/error.log")
	cmd2 := exec.Command("grep", fmt.Sprintf("%s:", tm))
	cmd3 := exec.Command("wc", "-l")
	cmds := []*exec.Cmd{cmd1, cmd2, cmd3}
	count, err := pipes.Run(cmds)
	if err != nil {
		return
	}

	v, _ := strconv.Atoi(count)
	m = v / 60
	return
}
