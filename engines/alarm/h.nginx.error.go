package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
	"github.com/qxnw/lib4go/transform"
)

//NginxErrorCountCollect nginx错误数收集
func NginxErrorCountCollect(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("max"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "nginx错误数")
		msg := ctx.Request.Setting.GetString("msg", "@host服务器nginx错误数:@current(@ct)")
		platform := ctx.Request.Setting.GetString("platform", "----")
		maxValue := ctx.Request.Setting.GetInt("max")
		current, ct, err := getNginxErrorCount()
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
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("ct", ct)
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, value, "nginx-error")
		response.SetContent(st, err)
		return
	}
}
func getNginxErrorCount() (m int, tm string, err error) {
	tm = time.Now().Add(-1 * time.Minute).Format("15:04")
	cmd := fmt.Sprintf(`cat /usr/local/nginx/logs/error.log|grep "%s:"|wc -l`, tm)
	count, err := pipes.BashRun(cmd)
	if err != nil {
		return
	}
	v, _ := strconv.Atoi(count)
	m = v
	return
}