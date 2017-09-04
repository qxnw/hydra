package rpc

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/security/jwt"
)

//JWTFilter jwt
func JWTFilter() HandlerFunc {
	return func(ctx *Context) {
		jwtAuth := ctx.server.jwt
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
		url := ctx.Req().GetService()
		for _, u := range jwtAuth.Exclude {
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
func (ctx *Context) SetJwtToken(data interface{}) {
	if data == nil {
		return
	}
	ctx.jwtStorage = data
	jwtAuth := ctx.server.jwt
	jwtToken, err := jwt.Encrypt(jwtAuth.Secret, jwtAuth.Mode, ctx.jwtStorage, jwtAuth.ExpireAt)
	if err != nil {
		ctx.WriteHeader(500)
		ctx.Result = &StatusResult{Code: 500, Result: fmt.Errorf("jwt配置出错：%v", err)}
		return
	}
	if ctx.Writer.Params == nil {
		ctx.Writer.Params = make(map[string]interface{})
	}
	ctx.Writer.Params[jwtAuth.Name] = jwtToken
}

// CheckJWT 检查jwk参数是否合法
func (ctx *Context) checkJWT(name string, secret string) (data interface{}, err context.Error) {
	token := ctx.getToken(name)
	if token == "" {
		return nil, context.NewError(403, fmt.Errorf("%s未传入jwt.token", name))
	}
	data, er := jwt.Decrypt(token, ctx.server.jwt.Secret)
	if er != nil {
		if strings.Contains(er.Error(), "Token is expired") {
			return nil, context.NewError(401, er)
		}
		return data, context.NewError(403, er)
	}
	return data, nil
}
