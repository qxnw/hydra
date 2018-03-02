package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/transform"
)

//DiskUPCollect 硬盘使用率收集
func DiskUPCollect(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("max"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "服务器disk使用率")
		msg := ctx.Request.Setting.GetString("msg", "@host服务器disk使用率:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		maxValue := ctx.Request.Setting.GetFloat64("max")
		diskInfo := disk.GetInfo()
		result := 1
		if diskInfo.UsedPercent < maxValue {
			result = 0
		}
		tf := transform.New()
		tf.Set("host", net.LocalIP)
		tf.Set("value", strconv.Itoa(result))
		tf.Set("current", fmt.Sprintf("%.2f", diskInfo.UsedPercent))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, result, "disk")
		response.SetContent(st, err)
		return
	}
}
