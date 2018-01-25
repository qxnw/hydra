package alarm

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

//HydraServerResponseCodeCollect hydra服务器的请求码收集
func HydraServerResponseCodeCollect(c component.IContainer, tp string) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("domain"); err != nil {
			response.SetStatus(500)
			return
		}

		title := ctx.Request.Setting.GetString("title", "请求响应码")
		msg := ctx.Request.Setting.GetString("msg", "@url请求响应码:@code在@span内出现:@current次")
		platform := ctx.Request.Setting.GetString("platform", "----")
		domain := ctx.Request.Setting.GetString("domain")
		max := ctx.Request.Setting.GetInt("max", 0)
		min := ctx.Request.Setting.GetInt("min", 0)

		tf := transform.New()
		tf.Set("domain", domain)
		tf.Set("span", "5m")
		tf.Set("code", ctx.Request.Setting.GetString("code", "500"))

		sql := tf.Translate(srvQueryMap[tp])
		urls, values, err := query(c, sql, tf)
		if err != nil {
			return
		}
		if len(urls) == 0 {
			response.SetStatus(204)
			return
		}
		for i, url := range urls {
			value := 1 //需要报警
			val := values[i]
			if ((min > 0 && val >= min) || min == 0) && ((max > 0 && val < max) || max == 0) {
				value = 0 //恢复
			}
			tf.Set("url", url)
			tf.Set("value", strconv.Itoa(value))
			tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
			tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
			tf.Set("current", strconv.Itoa(val))
			tf.Set("time", time.Now().Format("20060102150405"))
			tf.Set("unq", tf.Translate("{@domain}_{@url}_{@code}"))
			tf.Set("title", tf.TranslateAll(title, true))
			tf.Set("msg", tf.TranslateAll(msg, true))
			tf.Set("platform", platform)
			st, err := checkAndSave(c, tf, value, tp)
			if err != nil {
				response.SetError(st, err)
				return response, err
			}
		}
		return
	}
}
