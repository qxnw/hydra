package mock

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
)

func RawRequest(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		content, err := c.GetVarParam("setting", ctx.Request.Setting.GetString("setting"))
		if err != nil {
			err = fmt.Errorf("args配置错误，args.setting配置的节点获取失败(err:%v)", err)
			return
		}

		response.SetHeader("Content-Type", "text/plain")
		response.SetContent(200, ctx.Request.Translate(content, true))
		header := ctx.Request.Setting.GetString("header")
		if header == "" {
			return
		}
		headerContent, err := c.GetVarParam("header", header)
		if err != nil {
			err = fmt.Errorf("args配置错误，args.header配置的节点:%s获取失败(err:%v)", header, err)
			return
		}

		mapHeader, err := conf.NewJSONConfWithJson(headerContent, 0, nil)
		if err != nil {
			return
		}

		mapHeader.Each(func(k string) {
			response.SetHeader(k, ctx.Request.Translate(mapHeader.String(k), true))
		})
		return
	}
}
