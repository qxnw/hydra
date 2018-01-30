package middleware

import (
	"github.com/qxnw/hydra/servers/http"

	"github.com/gin-gonic/gin"
)

//AjaxRequest ajax请求限制
func AjaxRequest(conf *http.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if conf.OnlyAllowAjaxRequest && ctx.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			ctx.AbortWithStatus(510)
			return
		}
		ctx.Next()
	}
}
