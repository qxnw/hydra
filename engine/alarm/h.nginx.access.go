package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) nginxAccessCountCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	title := ctx.Input.GetArgsValue("title", "nginx 每分钟请求数")
	msg := ctx.Input.GetArgsValue("msg", "@host服务器 nginx每分钟请求数:@current(@ct)")
	platform := ctx.Input.GetArgsValue("platform", "----")
	maxValue, err := ctx.Input.GetArgsIntValue("max")
	if err != nil {
		return
	}
	current, ct, err := s.getNginxQPSCount()
	if err != nil {
		return
	}
	value := 1
	if current < maxValue {
		value = 0
	}
	tf := transform.New()
	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", strconv.Itoa(current))
	tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
	tf.Set("ct", ct)
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	tf.Set("platform", platform)
	st, err := s.checkAndSave(ctx, "nginx-access", tf, value)
	response.SetError(st, err)
	return
}
func (s *collectProxy) getNginxQPSCount() (m int, tm string, err error) {
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
