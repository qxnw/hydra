package monitor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

func (s *monitorProxy) nginxAccessCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	c, _, err := s.getNginxAccessCount()
	if err != nil {
		return
	}
	err = updateNginxAccessCount(ctx, int64(c), "server", ip)
	response.SetError(0, err)
	return
}
func (s *monitorProxy) getNginxAccessCount() (m int, tm string, err error) {
	tm = time.Now().Add(-1 * time.Minute).Format("15:04")
	cmd := fmt.Sprintf(`cat /usr/local/nginx/logs/access.log|grep "%s:"|wc -l`, tm)
	count, err := pipes.BashRun(cmd)
	if err != nil {
		return
	}
	v, _ := strconv.Atoi(count)
	m = v
	return
}
