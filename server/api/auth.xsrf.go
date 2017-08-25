package api

import "github.com/qxnw/lib4go/security/xsrf"

type Auth struct {
	Name     string
	ExpireAt int64
	Mode     string
	Secret   string
	Exclude  []string
	Enable   bool
}

func XSRFFilter() HandlerFunc {
	return func(ctx *Context) {
		if ctx.Server.xsrf == nil || !ctx.Server.xsrf.Enable || ctx.CheckXSRFToken(ctx.Server.xsrf.Name, ctx.Server.xsrf.Secret) {
			ctx.Next()
			return
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
