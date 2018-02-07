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
		response := getResponse(ctx)
		if response == nil {
			getLogger(ctx).Warn("response is nil")
			return
		}
		defer response.Close()
		if response.GetError() != nil {
			getLogger(ctx).Errorf("err:%v", response.GetError())
		}
		if ctx.Writer.Written() {
			return
		}
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
