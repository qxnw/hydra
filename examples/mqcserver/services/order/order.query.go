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
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	sdi, sdc := ctx.Request.Ext.GetSharding()
	ctx.Log.Infof("%s %d/%d", service, sdi, sdc)
	return "success"
}
