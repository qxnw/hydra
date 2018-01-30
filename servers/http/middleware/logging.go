package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/http"
	"github.com/qxnw/lib4go/logger"
)

//Logging 记录日志
func Logging(conf *http.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sid := ctx.Keys["hydra_sid"].(string)
		start := time.Now()
		p := ctx.Request.URL.Path
		if len(ctx.Request.URL.RawQuery) > 0 {
			p = p + "?" + ctx.Request.URL.RawQuery
		}
		log := logger.GetSession(conf.GetFullName(), sid)
		log.Info(conf.Type+".request:", conf.Name, ctx.Request.Method, p, "from", ctx.ClientIP())

		ctx.Next()

		statusCode := ctx.Writer.Status()

		if statusCode >= 200 && statusCode < 400 {
			log.Info(conf.Type+".response:", conf.Name, ctx.Request.Method, p, statusCode, time.Since(start))
		} else {
			log.Error(conf.Type+".response:", conf.Name, ctx.Request.Method, p, statusCode, time.Since(start))
		}
	}

}
