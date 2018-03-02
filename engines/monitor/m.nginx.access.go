package monitor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
)

//CollectNginxAccessNum 收集nginx访问数量
func CollectNginxAccessNum(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		ip := net.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		n, _, err := getNginxAccessCount()
		if err != nil {
			return
		}
		err = updateNginxAccessCount(c, ctx, int64(n), "server", ip)
		response.SetContent(0, err)
		return
	}
}
func getNginxAccessCount() (m int, tm string, err error) {
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
