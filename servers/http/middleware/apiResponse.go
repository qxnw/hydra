package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//APIResponse 处理api返回值
func APIResponse(conf *conf.ApiServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if ctx.Writer.Written() {
			return
		}
		response := getResponse(ctx)
		if response == nil {
			return
		}
		defer response.Close()
		switch response.GetContentType() {
		case 1:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		case 2:
			ctx.XML(response.GetStatus(), response.GetContent())
		case 3:
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		default:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		}
	}
}
