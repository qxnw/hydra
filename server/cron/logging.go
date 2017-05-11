package cron

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
	return func(ctx *Task) {
		start := time.Now()
		ctx.Info("cron.request:", ctx.server.serverName, "for", ctx.taskName)
		ctx.DoNext()
		ctx.Info("cron.response:", ctx.server.serverName, "for", ctx.taskName, time.Since(start), "status", ctx.statusCode)

	}
}
