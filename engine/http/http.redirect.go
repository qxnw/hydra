package http

import "github.com/qxnw/hydra/context"

func (s *httpProxy) httpRedirectHandle(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	err = ctx.Input.CheckArgs("url")
	if err != nil {
		return
	}
	code := ctx.Input.GetArgInt("status", 302)
	response.Redirect(code, ctx.Input.GetArgValue("url"))
	return

}