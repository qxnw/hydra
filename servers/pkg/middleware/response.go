package middleware

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//Response 处理api返回值
func Response(conf *conf.ServerConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		response := getResponse(ctx)
		if response == nil {
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
			if content, ok := response.GetContent().(string); ok {
				if (strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{")) &&
					(strings.HasSuffix(content, "}") || strings.HasSuffix(content, "]")) {
					ctx.SecureJSON(response.GetStatus(), response.GetContent())
				} else {
					ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
				}
				return
			}
			ctx.Data(response.GetStatus(), "text/plain", []byte(fmt.Sprint(response.GetContent())))
		}
	}
}
