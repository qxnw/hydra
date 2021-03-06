package user

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type LoginHandler struct {
	container component.IContainer
	Name      string
}

func NewLoginHandler(container component.IContainer) (u *LoginHandler) {
	return &LoginHandler{container: container, Name: "LoginHandler"}
}
func (u *LoginHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	//检查用户名密码是否正确
	ctx.Response.SetJWTBody(map[string]interface{}{
		"id": 11000,
	})
	return "ok"
}
