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
		span, err := ctx.GetArgByName("span")
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
		tf.Set("span", span)
		tf.Set("code", ctx.GetArgValue("code", "500"))

		sql := tf.Translate(s.srvQueryMap[tp])
		domains, values, err := s.query(ctx, sql, tf)
		if err != nil {
			return
		}
		for i, domain := range domains {
			value := 1 //需要报警
			val := values[i]
			if ((min > 0 && val >= min) || min == 0) && ((max > 0 && val < max) || max == 0) {
				value = 0 //恢复
			}
			tf.Set("domain", domain)
			tf.Set("value", strconv.Itoa(value))
			tf.Set("level", ctx.GetArgValue("level", "1"))
			tf.Set("group", ctx.GetArgValue("group", "D"))
			tf.Set("current", strconv.Itoa(val))
			tf.Set("time", time.Now().Format("20060102150405"))
			tf.Set("msg", tf.TranslateAll(ctx.GetArgValue("msg", "-"), true))
			st, err = s.checkAndSave(ctx, tp, tf, value)
			if err != nil {
				return
			}
		}
		return
	}
}
