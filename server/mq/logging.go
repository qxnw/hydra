package mq

import "time"

func Logging() HandlerFunc {
	return func(ctx *Context) {
		start := time.Now()
		ctx.Info("mq.request:", ctx.server.serverName, ctx.queue)
		ctx.Debugf("mq.request.raw:%s", ctx.msg.GetMessage())
		ctx.Next()
		if ctx.statusCode >= 200 && ctx.statusCode < 400 {
			ctx.Info("mq.response:", ctx.server.serverName, ctx.queue, ctx.statusCode, time.Since(start))
		} else {
			ctx.Error("mq.response:", ctx.server.serverName, ctx.queue, ctx.statusCode, time.Since(start))
		}
	}
}
