package middleware

import (
	"github.com/qxnw/hydra/servers/http"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/context"
)

//APIResponse 处理api返回值
func APIResponse(conf *http.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Writer.Written() {
			ctx.Next()
			return
		}
		result, ok := ctx.Get("__response_")
		if !ok || result == nil {
			ctx.Next()
			return
		}
		response, ok := result.(context.Response)
		if !ok || response == nil {
			ctx.Next()
			return
		}
		switch response.GetContentType() {
		case 1:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		case 2:
			ctx.XML(response.GetStatus(), response.GetContent())
		case 3:
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		}
		ctx.Next()
	}
}
