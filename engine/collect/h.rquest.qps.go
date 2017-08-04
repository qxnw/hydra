package collect

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) requestQPSCollect(tp string) func(ctx *context.Context) (r string, st int, err error) {

	return func(ctx *context.Context) (r string, st int, err error) {
		title := ctx.GetArgValue("title", "每秒钟请求数")
		msg := ctx.GetArgValue("msg", "@url在@span内请求:@current次")

		domain, err := ctx.GetArgByName("domain")
		if err != nil {
			return
		}
		max := types.ToInt(ctx.GetArgValue("max", "0"), 0)
		if err != nil {
			return
		}
		min := types.ToInt(ctx.GetArgValue("min", "0"), 0)
		if err != nil {
			return
		}
		tf := transform.New()
		tf.Set("domain", domain)
		tf.Set("span", "5m")

		sql := tf.Translate(s.srvQueryMap[tp])
		urls, values, err := s.query(ctx, sql, tf)
		if err != nil {
			return
		}
		if len(urls) == 0 {
			st = 204
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
			tf.Set("level", ctx.GetArgValue("level", "1"))
			tf.Set("group", ctx.GetArgValue("group", "D"))
			tf.Set("current", strconv.Itoa(val))
			tf.Set("time", time.Now().Format("20060102150405"))
			tf.Set("unq", tf.Translate("{@domain}_{@url}_QPS"))
			tf.Set("title", tf.TranslateAll(title, true))
			tf.Set("msg", tf.TranslateAll(msg, true))
			st, err = s.checkAndSave(ctx, tp, tf, value)
			if err != nil {
				return
			}
		}
		return
	}
}
