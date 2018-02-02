package middleware

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//Response 处理api返回值
func Response(conf *conf.ServerConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		if ctx.Writer.Written() {
			return
		}
		response := getResponse(ctx)
		if response == nil {
			return
		}
		defer response.Close()
		switch response.GetContentType() {
		case 1:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		case 2:
			ctx.XML(response.GetStatus(), response.GetContent())
		default:
			ctx.Data(response.GetStatus(), "text/plain", []byte(fmt.Sprint(response.GetContent())))
		}
	}
}
