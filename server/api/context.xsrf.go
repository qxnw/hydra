package api

import "github.com/qxnw/lib4go/security/xsrf"

type XSRF struct {
	Enable bool
	Key    string
	Secret string
}

// CheckXSRFToken 获取当前XSRFCookie
func (ctx *Context) CheckXSRFToken(key string, secret string) bool {
	xsrfToken := ""
	if cookie := ctx.Cookies().Get(key); cookie != nil {
		xsrfToken = cookie.Value
	}
	if xsrfToken == "" {
		xsrfToken = ctx.req.Header.Get(key)
	}
	if xsrfToken == "" {
		tk := ctx.Forms().Form[key]
		if len(tk) > 0 {
			xsrfToken = tk[0]
		}
	}
	if xsrfToken == "" {
		return false
	}

	v := xsrf.ParseXSRFToken(secret, xsrfToken)
	if v == "" {
		return false
	}
	return true
}

func XSRFFilter() HandlerFunc {
	return func(ctx *Context) {
		if ctx.Server.xsrf == nil || !ctx.Server.xsrf.Enable || ctx.CheckXSRFToken(ctx.Server.xsrf.Key, ctx.Server.xsrf.Secret) {
			ctx.Next()
			return
		}
		ctx.WriteHeader(403)
		ctx.Result = &StatusResult{Code: 403}
		return

	}
}
