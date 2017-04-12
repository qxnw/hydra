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
		ctx.Info("Started", ctx.taskName, "for", ctx.params)

		ctx.Next()

		if ctx.err == nil || ctx.statusCode == 200 {
			ctx.Info(ctx.taskName, ctx.statusCode, time.Since(start), ctx.Result)
		} else {
			ctx.Error(ctx.taskName, ctx.statusCode, time.Since(start), ctx.Result)
		}
	}
}
