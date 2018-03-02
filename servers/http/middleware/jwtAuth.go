package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/security/jwt"
)

//JwtAuth jwt
func JwtAuth(cnf *conf.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jwtAuth, ok := cnf.GetMetadata("jwt").(*conf.Auth)
		if !ok || jwtAuth == nil || !jwtAuth.Enable {
			ctx.Next()
			return
		}

		//检查jwt.token是否正确
		data, err := checkJWT(ctx, jwtAuth.Name, jwtAuth.Secret)
		if err == nil {
			setJWTRaw(ctx, data)
			ctx.Next()
			return
		}

		//不需要校验的URL自动跳过
		url := ctx.Request.URL.Path
		for _, u := range jwtAuth.Exclude {
			if u == url {
				ctx.Next()
				return
			}
		}
		//jwt.token错误，返回错误码
		ctx.AbortWithError(err.Code(), err)
		getLogger(ctx).Errorf("jwt:%v", err)
		return

	}
}
func setJwtResponse(ctx *gin.Context, cnf *conf.ServerConf, data interface{}) {
	if data == nil {
		data = getJWTRaw(ctx)
	}
	jwtAuth, ok := cnf.GetMetadata("jwt").(*conf.Auth)
	if !ok || !jwtAuth.Enable {
		return
	}
	jwtToken, err := jwt.Encrypt(jwtAuth.Secret, jwtAuth.Mode, data, jwtAuth.ExpireAt)
	if err != nil {
		ctx.AbortWithError(500, fmt.Errorf("jwt配置出错：%v", err))
		return
	}
	ctx.Header("Set-Cookie", fmt.Sprintf("%s=%s;path=/;", jwtAuth.Name, jwtToken))
}

// CheckJWT 检查jwk参数是否合法
func checkJWT(ctx *gin.Context, name string, secret string) (data interface{}, err context.Error) {
	token := getToken(ctx, name)
	if token == "" {
		return nil, context.NewError(403, fmt.Errorf("获取%s失败或未传入该参数", name))
	}
	data, er := jwt.Decrypt(token, secret)
	if er != nil {
		if strings.Contains(er.Error(), "Token is expired") {
			return nil, context.NewError(401, er)
		}
		return data, context.NewError(403, er)
	}
	return data, nil
}
func getToken(ctx *gin.Context, key string) string {
	if cookie, err := ctx.Cookie(key); err == nil {
		return cookie
	}
	return ""
}
func setToken(ctx *gin.Context, name string, token string) {
	ctx.Header(name, token)
}
