package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//Header 头设置
func Header(conf *conf.ServerConf) gin.HandlerFunc {
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
							ctx.Header(k, h)
							continue
						}
					}
				}
				continue
			}
			ctx.Header(k, v)
		}
		response := getResponse(ctx)
		if response == nil {
			return
		}
		header := response.GetHeaders()
		for k, v := range header {
			ctx.Header(k, v)
		}
	}
}
