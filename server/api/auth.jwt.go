package api

import "github.com/qxnw/lib4go/security/jwt"
import "fmt"

//JWTFilter jwt
func JWTFilter() HandlerFunc {
	return func(ctx *Context) {
		jwtAuth := ctx.Server.jwt
		if jwtAuth == nil || !jwtAuth.Enable {
			ctx.Next()
			return
		}

		//检查jwt.token是否正确
		data, err := ctx.checkJWT(jwtAuth.Name, jwtAuth.Secret)
		if err == nil {
			ctx.jwtStorage = data
			ctx.Next()
			return
		}

		//不需要校验的URL自动跳过
		url := ctx.Req().RequestURI
		for _, u := range jwtAuth.Exclude {
			if u == url {
				ctx.Next()
				return
			}
		}

		//jwt.token错误，返回错误码
		ctx.WriteHeader(403)
		ctx.Result = &StatusResult{Code: 403, Result: err}
		ctx.Error(err)
		return

	}
}
func (ctx *Context) setJwtToken(data interface{}) {
	if data == nil {
		return
	}
	ctx.jwtStorage = data
	jwtAuth := ctx.Server.jwt
	jwtToken, err := jwt.Encrypt(jwtAuth.Secret, jwtAuth.Mode, ctx.jwtStorage, jwtAuth.ExpireAt)
	if err != nil {
		ctx.WriteHeader(500)
		ctx.Result = &StatusResult{Code: 500, Result: fmt.Errorf("jwt配置出错：%v", err)}
		return
	}
	ctx.Header().Set(jwtAuth.Name, jwtToken)
}

// CheckJWT 检查jwk参数是否合法
func (ctx *Context) checkJWT(name string, secret string) (data interface{}, err error) {
	token := ctx.getToken(name)
	if token == "" {
		return "", fmt.Errorf("%s未传入jwt.token", name)
	}
	return jwt.Decrypt(token, ctx.Server.jwt.Secret)
}
