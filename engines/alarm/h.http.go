package alarm

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

//HTTPStatusCollect http状态收集
func HTTPStatusCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("url"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "HTTP服务器")
		msg := ctx.Request.Setting.GetString("msg", "HTTP服务器地址:@url请求响应码:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		uri := ctx.Request.Setting.GetString("url")
		u, err := url.Parse(uri)
		if err != nil {
			err = fmt.Errorf("http请求参数url配置有误:%v", uri)
			return
		}
		client := http.NewHTTPClient()
		_, t, err := client.Get(uri)
		value := types.DecodeInt(t, 200, 0, 1)
		tf := transform.New()
		tf.Set("host", u.Host)
		tf.Set("url", uri)
		tf.Set("value", strconv.Itoa(value))
		tf.Set("current", strconv.Itoa(t))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@url"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, value, "http")
		response.SetContent(st, err)
		return
	}
}
