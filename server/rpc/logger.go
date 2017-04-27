package rpc

import (
	"fmt"
	"io"
	"time"

	"github.com/lunny/log"
	"github.com/qxnw/lib4go/logger"
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
}

type nLogger struct {
	*log.Logger
}

func (n *nLogger) Fatalln(args ...interface{}) {
	n.Fatal(args...)
}

func NewLogger(name string, out io.Writer) Logger {
	l := log.New(out, "["+name+"] ", log.Ldefault())
	l.SetOutputLevel(log.Ldebug)
	return &nLogger{Logger: l}
}

type LogInterface interface {
	SetLogger(*logger.Logger)
}

type Log struct {
	*logger.Logger
}

func (l *Log) SetLogger(log *logger.Logger) {
	l.Logger = log
}

func Logging() HandlerFunc {
	fmt.Println("loggin....")
	return func(ctx *Context) {
		start := time.Now()
		ctx.Info("req.rpc", ctx.server.serverName, "for", ctx.Req().Service)
		ctx.Next()
		ctx.Info("res.rpc", ctx.server.serverName, "for", ctx.Req().Service, time.Since(start), "status", ctx.GetStatusCode())
	}
}
