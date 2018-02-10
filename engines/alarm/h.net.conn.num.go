package alarm

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
	"github.com/qxnw/lib4go/transform"
)

//NetConnNumCollect 网络连接数收集
func NetConnNumCollect(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("max"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "网络连接数")
		platform := ctx.Request.Setting.GetString("platform", "----")
		msg := ctx.Request.Setting.GetString("msg", "@host服务器网络连接数:@current")
		maxValue := ctx.Request.Setting.GetInt("max")
		ncc, err := getNetConnectNum()
		if err != nil {
			return
		}
		value := 1
		if ncc < maxValue {
			value = 0
		}
		tf := transform.New()
		tf.Set("host", net.LocalIP)
		tf.Set("value", strconv.Itoa(value))
		tf.Set("current", strconv.Itoa(ncc))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, value, "ncc")
		response.SetContent(st, err)
		return
	}
}

func getNetConnectNum() (v int, err error) {
	count, err := pipes.BashRun(`netstat -an|grep tcp|wc -l`)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
