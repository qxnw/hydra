package rpc

import "github.com/qxnw/lib4go/security/xsrf"

func XSRFFilter() HandlerFunc {
	return func(ctx *Context) {
		xsrfAuth := ctx.server.xsrf
		if xsrfAuth == nil || !xsrfAuth.Enable {
			ctx.Next()
			return
		}
		if ctx.CheckXSRFToken(xsrfAuth.Name, xsrfAuth.Secret) {
			ctx.Next()
			return
		}
		//不需要校验的URL自动跳过
		url := ctx.Req().GetService()
		for _, u := range xsrfAuth.Exclude {
			if u == url {
				ctx.Next()
				return
			}
		}
		ctx.WriteHeader(403)
		ctx.Result = &StatusResult{Code: 403}
		return

	}
}

// CheckXSRFToken 获取当前XSRFCookie
func (ctx *Context) CheckXSRFToken(name string, secret string) bool {
	xsrfToken := ctx.getToken(name)
	if xsrfToken == "" {
		return false
	}
	v := xsrf.ParseXSRFToken(secret, xsrfToken)
	if v == "" {
		return false
	}
	return true
}
