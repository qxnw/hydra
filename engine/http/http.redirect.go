package http

import (
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/goplugin"
)

func (s *httpProxy) httpRedirectHandle(ctx *context.Context) (r string, t int, param map[string]interface{}, err error) {
	context, err := goplugin.GetContext(ctx, s.ctx.Invoker)
	if err != nil {
		return
	}
	err = context.CheckArgs("url")
	if err != nil {
		return
	}
	code := context.GetArgsValue("status", "302")
	param = make(map[string]interface{})
	param["Status"] = code
	param["Location"] = context.GetArgsValue("url")
	return

}
