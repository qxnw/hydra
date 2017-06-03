package cron

import "time"

func Logging() HandlerFunc {
	return func(ctx *Task) {
		start := time.Now()
		ctx.Info("cron.request:", ctx.server.serverName, ctx.taskName)
		ctx.DoNext()
		if ctx.statusCode == 200 {
			ctx.Info("cron.response:", ctx.server.serverName, ctx.taskName, ctx.statusCode, time.Since(start))

		} else {
			ctx.Error("cron.response:", ctx.server.serverName, ctx.taskName, ctx.statusCode, time.Since(start))

		}

	}
}
