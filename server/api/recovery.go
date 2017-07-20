// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

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
				fmt.Fprintf(&buf, "api request handler crashed with error: %v", e)
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
				ctx.Logger.Error(content)
				if len(ctx.tan.headers) > 0 {
					for k, v := range ctx.tan.headers {
						ctx.Header().Set(k, v)
					}
				}
				if !ctx.Written() {
					if !debug {
						ctx.Result = InternalServerError(http.StatusText(http.StatusInternalServerError))
					} else {
						ctx.Result = InternalServerError(e.(error).Error())
					}
				}
			}
		}()

		ctx.Next()
	}
}
