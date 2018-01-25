package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/transform"
)

//CPUUPCollect CPU使用率收集
func CPUUPCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("max"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "服务器CPU负载")
		msg := ctx.Request.Setting.GetString("msg", "@host服务器CPU负载:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		maxValue := ctx.Request.Setting.GetFloat64("max")
		cpuInfo := cpu.GetInfo()
		value := 1
		if cpuInfo.UsedPercent < maxValue {
			value = 0
		}
		tf := transform.New()

		tf.Set("host", net.LocalIP)
		tf.Set("value", strconv.Itoa(value))
		tf.Set("current", fmt.Sprintf("%.2f", cpuInfo.UsedPercent))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, value, "cpu")
		response.SetError(st, err)
		return
	}
}
