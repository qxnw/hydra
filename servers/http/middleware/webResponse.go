package middleware

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/servers/pkg/conf"

	"github.com/gin-gonic/gin"
)

//WebResponse 处理web返回值
func WebResponse(conf *conf.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		case 3:
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		default:
			renderHtml(ctx, response)
		}
	}
}
func renderHtml(ctx *gin.Context, response context.Response) {
	defer func() {
		if err := recover(); err != nil {
			getLogger(ctx).Error(err)
		}
	}()
	names := strings.Split(getServiceName(ctx), "/")
	viewName := fmt.Sprintf("%s.html", names[len(names)-1])
	ctx.HTML(response.GetStatus(), viewName, response.GetContent())
}
