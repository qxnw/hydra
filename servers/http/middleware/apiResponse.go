package middleware

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
)

//APIResponse 处理api返回值
func APIResponse(conf *conf.MetadataConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		nctx := getCTX(ctx)
		if nctx == nil {
			return
		}
		defer nctx.Close()
		if err := nctx.Response.GetError(); err != nil {
			getLogger(ctx).Error(err)
			ctx.AbortWithStatus(nctx.Response.GetStatus())
			return
		}
		if ctx.Writer.Written() {
			return
		}
		switch nctx.Response.GetContentType() {
		case context.CT_XML:
			ctx.XML(nctx.Response.GetStatus(), nctx.Response.GetContent())
		case context.CT_YMAL:
			ctx.YAML(nctx.Response.GetStatus(), nctx.Response.GetContent())
		case context.CT_PLAIN:
			ctx.Data(nctx.Response.GetStatus(), "text/plain", []byte(fmt.Sprint(nctx.Response.GetContent())))
		case context.CT_HTML:
			ctx.Data(nctx.Response.GetStatus(), "text/html", []byte(fmt.Sprint(nctx.Response.GetContent())))
		default:
			ctx.SecureJSON(nctx.Response.GetStatus(), getJsonMessage(nctx.Response.GetContent()))
		}
	}
}
func getJsonMessage(i interface{}) interface{} {
	switch i.(type) {
	case string:
		return json.RawMessage(i.(string))
	default:
		return i
	}
}
