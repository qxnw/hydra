package product

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type QueryHandler struct {
	fields    map[string][]string
	container component.IContainer
}

func NewQueryHandler(container component.IContainer) (u *QueryHandler) {
	u = &QueryHandler{
		container: container,
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	return
}

func (u *QueryHandler) Handle(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	ctx.Log.Info("-------order.query-------")
	sharding_index, sharding_count := ctx.Request.Ext.GetSharding()
	ctx.Log.Infof("sharding:index-%d,count-%d", sharding_index, sharding_count)
	response.Success("success")
	return
}

func (u *QueryHandler) Close() error {
	return nil
}
