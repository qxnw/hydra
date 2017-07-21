package api

func OnlyAllowAjaxRequest() HandlerFunc {
	return func(ctx *Context) {
		if ctx.Server.onlyAllowAjaxRequest && !ctx.IsAjax() {
			ctx.WriteHeader(4031)
			ctx.Result = &StatusResult{Code: 4031}
			return
		}
		ctx.Next()
	}
}
