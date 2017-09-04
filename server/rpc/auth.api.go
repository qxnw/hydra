package rpc

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/security/md5"
)

func (ctx *Context) checkAPIAuth(secret string) (err context.Error) {
	apiAuth := ctx.server.api
	if apiAuth == nil || !apiAuth.Enable {
		return context.NewError(context.ERR_NOT_EXTENDED, "api认证未配置")
	}
	return ctx.checkBasicAuth(apiAuth.ExpireAt, apiAuth.Secret)
}
func (ctx *Context) checkBasicAuth(expireAt int64, secret string) (err context.Error) {
	if secret == "" {
		return context.NewError(context.ERR_NOT_EXTENDED, errors.New("secret未配置"))
	}
	forms := ctx.Req().GetArgs()
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
	if sign == "" || timestamp == "" {
		return context.NewError(context.ERR_NOT_ACCEPTABLE, errors.New("缺少参数sign或timestamp"))
	}
	ts, er := time.Parse("20060102150405", timestamp)
	if er != nil {
		return context.NewError(context.ERR_FORBIDDEN, errors.New("timestamp格式错误"))
	}
	if expireAt != 0 && time.Now().Add(time.Second*time.Duration(expireAt)).After(ts) {
		return context.NewError(context.ERR_UNAUTHORIZED, errors.New("Sign is expired"))
	}
	sort.Strings(data)
	raw := fmt.Sprintf("%s%s%s", secret, strings.Join(data, ""), secret)
	nsign := md5.Encrypt(raw)
	if sign != nsign {
		return context.NewError(context.ERR_FORBIDDEN, fmt.Errorf("参数签名错误:%s", strings.Join(data, "")))
	}
	return nil
}
