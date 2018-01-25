package mock

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
)

func RawRequest() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		content, err := ctx.Request.Ext.GetVarParam("setting", ctx.Request.Setting.GetString("setting"))
		if err != nil {
			err = fmt.Errorf("args配置错误，args.setting配置的节点获取失败(err:%v)", err)
			return
		}
		paraTransform := transform.New()
		ctx.Request.Param.Each(func(k, v string) {
			paraTransform.Set(k, v)
		})
		ctx.Request.Form.Each(func(k, v string) {
			paraTransform.Set(k, v)
		})

		response.SetHeader("Content-Type", "text/plain")
		response.SetContent(200, paraTransform.Translate(content))
		header := ctx.Request.Setting.GetString("header")
		if header == "" {
			return
		}
		headerContent, err := ctx.Request.Ext.GetVarParam("header", header)
		if err != nil {
			err = fmt.Errorf("args配置错误，args.header配置的节点:%s获取失败(err:%v)", header, err)
			return
		}

		mapHeader, err := conf.NewJSONConfWithJson(headerContent, 0, nil)
		if err != nil {
			return
		}

		mapHeader.Each(func(k string) {
			response.SetHeader(k, paraTransform.Translate(mapHeader.String(k)))
		})
		return
	}
}
