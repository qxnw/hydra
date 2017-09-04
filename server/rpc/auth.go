package rpc

type Auth struct {
	Name     string
	ExpireAt int64
	Mode     string
	Secret   string
	Exclude  []string
	Enable   bool
}

// GetToken 获取安全验证token
func (ctx *Context) getToken(key string) string {
	token := ctx.Req().GetArgs()[key]
	return token
}
func (ctx *Context) setToken(name string, token string) {
	if ctx.Writer.Params == nil {
		ctx.Writer.Params = make(map[string]interface{})
	}
	ctx.Writer.Params[name] = token
}
