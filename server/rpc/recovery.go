// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
)

func Recovery(debug bool) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if e := recover(); e != nil {
				var buf bytes.Buffer
				fmt.Fprintf(&buf, "rpc request handler crashed with error: %v", e)

				for i := 1; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					} else {
						fmt.Fprintf(&buf, "\n")
					}
					fmt.Fprintf(&buf, "%v:%v", file, line)
				}
				var content = buf.String()
				ctx.Error(content)
				ctx.Writer.Code = 500
				if !ctx.Written() {
					if !debug {
						ctx.Result = &StatusResult{Code: 500, Result: InternalServerError(http.StatusText(http.StatusInternalServerError))}
					} else {
						ctx.Result = &StatusResult{Code: 500, Result: content}
					}
				}
			}
		}()

		ctx.Next()
	}
}
