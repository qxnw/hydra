package rpc

import (
	"io"
	"time"

	"github.com/lunny/log"
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
	Infof(format string, v ...interface{})
	Info(v ...interface{})
	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
	Errorf(format string, v ...interface{})
	Error(v ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

type nLogger struct {
	*log.Logger
}

func (n *nLogger) Fatalln(args ...interface{}) {
	n.Fatal(args...)
}

func NewLogger(out io.Writer) Logger {
	l := log.New(out, "[rpc] ", log.Ldefault())
	l.SetOutputLevel(log.Ldebug)
	return &nLogger{Logger: l}
}

type LogInterface interface {
	SetLogger(Logger)
}

type Log struct {
	Logger
}

func (l *Log) SetLogger(log Logger) {
	l.Logger = log
}

func Logging() HandlerFunc {
	return func(ctx *Context) {
		start := time.Now()
		ctx.server.logger.Info("Started", ctx.Req().Service, "for", ctx.Req().GetArgs()["session"])

		if action := ctx.Action(); action != nil {
			if l, ok := action.(LogInterface); ok {
				l.SetLogger(ctx.Logger)
			}
		}

		ctx.Next()

		if !ctx.Written() {
			if ctx.Result == nil {
				ctx.Result = NotFound()
			}
			ctx.HandleError()
		}

		statusCode := ctx.Writer.Code
		if statusCode >= 200 && statusCode < 400 {
			ctx.server.logger.Info(ctx.Req().Service, statusCode, time.Since(start), ctx.Result)
		} else {
			ctx.server.logger.Error(ctx.Req().Service, statusCode, time.Since(start), ctx.Result)
		}
	}
}
