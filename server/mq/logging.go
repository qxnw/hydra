package mq

import (
	"io"
	"time"

	"github.com/lunny/log"
	"github.com/qxnw/hydra/context"
)

type nLogger struct {
	*log.Logger
}

func (n *nLogger) Fatalln(args ...interface{}) {
	n.Fatal(args...)
}

func NewLogger(name string, out io.Writer) context.Logger {
	l := log.New(out, "["+name+"] ", log.Ldefault())
	l.SetOutputLevel(log.Ldebug)
	return &nLogger{Logger: l}
}
func Logging() HandlerFunc {
	return func(ctx *Context) {
		start := time.Now()
		ctx.Info("mq.request:", ctx.server.serverName, "for", ctx.queue)
		ctx.Next()
		if ctx.statusCode == 200 {
			ctx.Info("mq.response:", ctx.server.serverName, "for", ctx.queue, time.Since(start), "status", ctx.statusCode)
		} else {
			ctx.Error("mq.response:", ctx.server.serverName, "for", ctx.queue, time.Since(start), "status", ctx.statusCode)
		}
	}
}
