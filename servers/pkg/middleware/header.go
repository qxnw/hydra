package middleware

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//Header 头设置
func Header(conf *conf.ServerConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		for k, v := range conf.Headers {
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
