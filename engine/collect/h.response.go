package collect

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) responseCollect(tp string) context.HandlerFunc {

	return func(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
		response = context.GetResponse()
		title := ctx.Input.GetArgValue("title", "请求响应码")
		msg := ctx.Input.GetArgValue("msg", "@url请求响应码:@code在@span内出现:@current次")

		domain, err := ctx.Input.GetArgByName("domain")
		if err != nil {
			return
		}
		max := ctx.Input.GetArgInt("max", 0)
		min := ctx.Input.GetArgInt("min", 0)

		tf := transform.New()
		tf.Set("domain", domain)
		tf.Set("span", "5m")
		tf.Set("code", ctx.Input.GetArgValue("code", "500"))

		sql := tf.Translate(s.srvQueryMap[tp])
		urls, values, err := s.query(ctx, sql, tf)
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
			tf.Set("level", ctx.Input.GetArgValue("level", "1"))
			tf.Set("group", ctx.Input.GetArgValue("group", "D"))
			tf.Set("current", strconv.Itoa(val))
			tf.Set("time", time.Now().Format("20060102150405"))
			tf.Set("unq", tf.Translate("{@domain}_{@url}_{@code}"))
			tf.Set("title", tf.TranslateAll(title, true))
			tf.Set("msg", tf.TranslateAll(msg, true))
			st, err := s.checkAndSave(ctx, tp, tf, value)
			if err != nil {
				response.SetError(st, err)
				return response, err
			}
		}
		return
	}
}
