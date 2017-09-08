package api

import "strings"

func WriteHeader() HandlerFunc {
	return func(ctx *Context) {
		if len(ctx.Server.Headers) > 0 {
			for k, v := range ctx.Server.Headers {
				if k == "Access-Control-Allow-Origin" {
					if strings.Contains(v, ctx.req.Host) {
						hosts := strings.Split(v, ",")
						for _, h := range hosts {
							if strings.Contains(h, ctx.req.Host) {
								ctx.Header().Set(k, v)
								continue
							}
						}
					}
				}
				ctx.Header().Set(k, v)
			}
		}
		ctx.Next()
	}
}
