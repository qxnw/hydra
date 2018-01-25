package http

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//Redirect 请求转跳
func Redirect() component.WebServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.WebResponse, err error) {
		response = context.GetWebResponse(ctx)
		err = ctx.Request.Setting.Check("url")
		if err != nil {
			return
		}
		code := ctx.Request.Setting.GetInt("status", 302)
		response.Redirect(code, ctx.Request.Setting.GetString("url"))
		return
	}
}
