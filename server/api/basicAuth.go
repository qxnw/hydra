package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/security/md5"
)

func basicAuth() HandlerFunc {
	return func(ctx *Context) {
		forms := ctx.Forms().Form
		data := make([]string, 0, len(forms))
		sign := ""
		timestamp := ""
		for k, v := range forms {
			switch k {
			case "sign":
				sign = fmt.Sprint(v)
			case "timestamp":
				timestamp = fmt.Sprint(v)
				data = append(data, fmt.Sprint(v))
			default:
				data = append(data, fmt.Sprint(v))
			}
		}
		if sign == "" {
			ctx.WriteHeader(context.ERR_NOT_ACCEPTABLE)
			ctx.Result = &StatusResult{Code: context.ERR_NOT_ACCEPTABLE, Result: "缺少参数sign"}
			return
		}
		if timestamp == "" {
			ctx.WriteHeader(context.ERR_NOT_ACCEPTABLE)
			ctx.Result = &StatusResult{Code: context.ERR_NOT_ACCEPTABLE, Result: "缺少参数timestamp"}
			return
		}
		sort.Strings(data)
		raw := fmt.Sprintf("%s%s%s", ctx.Server.baseAuthSecret, strings.Join(data, ""), ctx.Server.baseAuthSecret)
		nsign := md5.Encrypt(raw)
		if sign != nsign {
			ctx.Debug("raw:", raw)
			ctx.WriteHeader(context.ERR_FORBIDDEN)
			ctx.Result = &StatusResult{Code: context.ERR_FORBIDDEN, Result: "参数签名错误"}
			return
		}
		ctx.Next()
	}
}
