// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"time"

	"reflect"

	"github.com/qxnw/lib4go/logger"
)

type LogInterface interface {
	SetLogger(*logger.Logger)
}

type Log struct {
	logger.Logger
}

func (l *Log) SetLogger(log logger.Logger) {
	l.Logger = log
}

func Logging() HandlerFunc {
	return func(ctx *Context) {
		start := time.Now()
		p := ctx.Req().URL.Path
		if len(ctx.Req().URL.RawQuery) > 0 {
			p = p + "?" + ctx.Req().URL.RawQuery
		}

		ctx.Info("api.request:", ctx.tan.serverName, ctx.Req().Method, p, "from", ctx.IP())

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

		statusCode := ctx.Status()

		if statusCode >= 200 && statusCode < 400 {
			ctx.Info("api.response:", ctx.tan.serverName, ctx.Req().Method, p, statusCode, time.Since(start))
		} else {
			if reflect.TypeOf(ctx.Result).Kind() == reflect.String {
				ctx.Error(ctx.Result.(string))
			}
			ctx.Error("api.response:", ctx.tan.serverName, ctx.Req().Method, p, statusCode, time.Since(start))
		}
	}
}
