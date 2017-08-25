package api

// GetToken 获取安全验证token
func (ctx *Context) getToken(key string) string {
	token := ctx.Req().Header.Get(key)
	if token != "" {
		return token
	}
	if cookie := ctx.Cookies().Get(key); cookie != nil {
		token = cookie.Value
	}
	if token != "" {
		return token
	}

	tk := ctx.Forms().Form[key]
	if len(tk) > 0 {
		token = tk[0]
	}

	return token
}
func (ctx *Context) setToken(name string, token string) {
	ctx.Header().Set(name, token)
}
