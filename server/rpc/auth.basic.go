package rpc

func basicAuth() HandlerFunc {
	return func(ctx *Context) {
		basicAuth := ctx.server.basic
		if basicAuth == nil || !basicAuth.Enable {
			ctx.Next()
			return
		}
		//检查basic.sign是否正确
		err := ctx.checkBasicAuth(basicAuth.ExpireAt, basicAuth.Secret)
		if err == nil {
			ctx.Next()
			return
		}

		//不需要校验的URL自动跳过
		url := ctx.Req().GetService()
		for _, u := range basicAuth.Exclude {
			if u == url {
				ctx.Next()
				return
			}
		}
		//jwt.token错误，返回错误码
		ctx.WriteHeader(err.Code())
		ctx.Result = &StatusResult{Code: err.Code(), Result: err}
		ctx.Error(err)
		return
	}
}
