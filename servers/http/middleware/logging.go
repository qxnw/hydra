package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/logger"
)

//Logging 记录日志
func Logging(conf *conf.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		setStartTime(ctx)
		p := ctx.Request.URL.Path
		if ctx.Request.URL.RawQuery != "" {
			p = p + "?" + ctx.Request.URL.RawQuery
		}
		uuid := getUUID(ctx)
		setUUID(ctx, uuid)
		log := logger.GetSession(conf.GetFullName(), uuid)
		log.Info(conf.Type+".request:", conf.Name, ctx.Request.Method, p, "from", ctx.ClientIP())
		setLogger(ctx, log)
		ctx.Next()

		statusCode := ctx.Writer.Status()
		if statusCode >= 200 && statusCode < 400 {
			log.Info(conf.Type+".response:", conf.Name, ctx.Request.Method, p, statusCode, getExt(ctx), time.Since(start))
		} else {
			log.Error(conf.Type+".response:", conf.Name, ctx.Request.Method, p, statusCode, getExt(ctx), time.Since(start))
		}
	}

}
