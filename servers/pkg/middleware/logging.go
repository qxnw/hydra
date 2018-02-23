package middleware

import (
	"time"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
	"github.com/qxnw/lib4go/logger"
)

//Logging 记录日志
func Logging(conf *conf.ServerConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		start := time.Now()
		setStartTime(ctx)
		p := ctx.Request.GetService()
		uuid := getUUID(ctx)
		setUUID(ctx, uuid)
		log := logger.GetSession(conf.GetFullName(), uuid)
		log.Info(conf.Type+".request:", conf.Name, ctx.Request.GetMethod(), p, "from", ctx.ClientIP())
		setLogger(ctx, log)
		ctx.Next()

		statusCode := ctx.Writer.Status()
		if statusCode >= 200 && statusCode < 400 {
			log.Info(conf.Type+".response:", conf.Name, ctx.Request.GetMethod(), p, statusCode, time.Since(start))
		} else {
			log.Error(conf.Type+".response:", conf.Name, ctx.Request.GetMethod(), p, statusCode, time.Since(start))
		}
	}

}
