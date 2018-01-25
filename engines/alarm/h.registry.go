package alarm

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

//RegistryNodeCountCollect 注册中心的节点数收集
func RegistryNodeCountCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("path", "min"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "注册中心服务")
		msg := ctx.Request.Setting.GetString("msg", "注册中心服务:@url当前数量:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		path := ctx.Request.Setting.GetString("path")
		minValue := ctx.Request.Setting.GetInt("min")
		data, _, err := c.GetRegistry().GetChildren(path)
		if err != nil {
			return
		}
		value := 1
		if len(data) >= minValue {
			value = 0
		}
		tf := transform.NewMap(map[string]string{})
		tf.Set("host", path)
		tf.Set("url", path)
		tf.Set("value", strconv.Itoa(value))
		tf.Set("current", strconv.Itoa(len(data)))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, value, "registry")
		response.SetError(st, err)
		return
	}
}
