package middleware

import (
	"fmt"
	"path/filepath"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/servers/pkg/conf"

	"github.com/gin-gonic/gin"
)

//WebResponse 处理web返回值
func WebResponse(conf *conf.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.Next()
		response := getResponse(ctx)
		if response == nil {
			return
		}
		defer response.Close()
		if response.GetError() != nil {
			getLogger(ctx).Error(response.GetError())
			ctx.AbortWithError(response.GetStatus(), response.GetError())
			return
		}
		if ctx.Writer.Written() {
			return
		}
		switch response.GetContentType() {
		case 1:
			ctx.SecureJSON(response.GetStatus(), response.GetContent())
		case 2:
			ctx.XML(response.GetStatus(), response.GetContent())
		case 3:
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		default:
			if renderHTML(ctx, response, conf) {
				return
			}
			ctx.Data(response.GetStatus(), "text/plain", []byte(response.GetContent().(string)))
		}
	}
}
func renderHTML(ctx *gin.Context, response context.Response, cnf *conf.ServerConf) bool {
	files, ok := cnf.GetMetadata("viewFiles").([]string)
	if !ok {
		return false
	}
	root := cnf.GetMetadata("view").(*conf.View).Path
	viewPath := filepath.Join(root, fmt.Sprintf("%s.html", getServiceName(ctx)))
	for _, f := range files {
		if f == viewPath {
			ctx.HTML(response.GetStatus(), filepath.Base(viewPath), response.GetContent())
			return true
		}
	}
	return false
}
