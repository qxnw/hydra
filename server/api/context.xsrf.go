package api

import "github.com/qxnw/lib4go/security/xsrf"

type XSRF struct {
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
		if ctx.tan.xsrf != nil && !ctx.CheckXSRFToken(ctx.tan.xsrf.Key, ctx.tan.xsrf.Secret) {
			ctx.WriteHeader(403)
			ctx.Result = &StatusResult{Code: 403}
			return
		}
		ctx.Next()
	}
}
