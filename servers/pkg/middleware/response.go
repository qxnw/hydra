package middleware

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//Response 处理api返回值
func Response(conf *conf.MetadataConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		context := getCTX(ctx)
		if context == nil {
			return
		}
		defer context.Close()
		if err := context.Response.GetError(); err != nil {
			getLogger(ctx).Errorf("err:%v", err)
		}
		if ctx.Writer.Written() {
			return
		}
		switch context.Response.GetContentType() {
		case 1:
			ctx.SecureJSON(context.Response.GetStatus(), context.Response.GetContent())
		case 2:
			ctx.XML(context.Response.GetStatus(), context.Response.GetContent())
		default:
			if content, ok := context.Response.GetContent().(string); ok {
				if (strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{")) &&
					(strings.HasSuffix(content, "}") || strings.HasSuffix(content, "]")) {
					ctx.SecureJSON(context.Response.GetStatus(), context.Response.GetContent())
				} else {
					ctx.Data(context.Response.GetStatus(), "text/plain", []byte(context.Response.GetContent().(string)))
				}
				return
			}
			ctx.Data(context.Response.GetStatus(), "text/plain", []byte(fmt.Sprint(context.Response.GetContent())))
		}
	}
}
