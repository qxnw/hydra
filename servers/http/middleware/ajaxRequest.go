package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//AjaxRequest ajax请求限制
func AjaxRequest(conf *conf.ApiServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if conf.OnlyAllowAjaxRequest && ctx.GetHeader("X-Requested-With") != "XMLHttpRequest" {
			ctx.AbortWithStatus(403)
			return
		}
		ctx.Next()
		return
	}
}
