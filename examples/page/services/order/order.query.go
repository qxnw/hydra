package order

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type QueryHandler struct {
	container component.IContainer
}

func NewQueryHandler(container component.IContainer) (u *QueryHandler) {
	return &QueryHandler{container: container}
}
func (u *QueryHandler) GetHandle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	return "get.success" + ctx.Request.Setting.GetString("db")
}
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	return "success"
}
