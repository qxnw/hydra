package middleware

import (
	"strings"

	"github.com/qxnw/hydra/servers/http"

	"github.com/gin-gonic/gin"
)

//Header 头设置
func Header(conf *http.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(conf.Headers) == 0 {
			return
		}
		for k, v := range conf.Headers {
			if k == "Access-Control-Allow-Origin" {
				if strings.Contains(v, ctx.Request.Host) {
					hosts := strings.Split(v, ",")
					for _, h := range hosts {
						if strings.Contains(h, ctx.Request.Host) {
							ctx.Header(k, v)
							continue
						}
					}
				}
			}
			ctx.Header(k, v)
		}

	}
}
