package middleware

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//NoResponse 处理无响应的返回结果
func NoResponse(conf *conf.ServerConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		response := getResponse(ctx)
		if response == nil {
			return
		}
		defer response.Close()
		if response.GetError() != nil {
			getLogger(ctx).Error(response.GetError())
		}
		if ctx.Writer.Written() {
			return
		}
		ctx.Writer.WriteHeader(response.GetStatus())
		ctx.Writer.WriteString(fmt.Sprint(response.GetContent()))
	}
}
