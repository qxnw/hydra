package mq

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
)

func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if e := recover(); e != nil {
				var buf bytes.Buffer
				fmt.Fprintf(&buf, "Handler crashed with error: %v", e)

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
				ctx.server.logger.Error(content)
				ctx.statusCode = http.StatusInternalServerError
				ctx.err = fmt.Errorf("%v", e)
				ctx.Result = ctx.err

			}
		}()

		ctx.Next()
	}
}
