package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//APIResponse 处理api返回值
func APIResponse(conf *conf.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		response := getResponse(ctx)
		if response == nil {
			return
		}
		defer response.Close()
		if response.GetError() != nil {
			getLogger(ctx).Error(response.GetError())
			ctx.AbortWithError(response.GetStatus(), response.GetError())
			return
		}
		if ctx.Writer.Written() {
			return
		}
		switch response.GetContentType() {
		case 1:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		case 2:
			ctx.XML(response.GetStatus(), response.GetContent())
		case 3:
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		default:
			if content, ok := response.GetContent().(string); ok {
				if (strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{")) &&
					(strings.HasSuffix(content, "}") || strings.HasSuffix(content, "]")) {
					ctx.SecureJSON(response.GetStatus(), response.GetContent())
				} else {
					ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
				}
				return
			}
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		}
	}
}
