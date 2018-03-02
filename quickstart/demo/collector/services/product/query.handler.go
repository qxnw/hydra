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
			"input": []string{"idxxx"},
		},
	}
	return
}
func (u *QueryHandler) GetFallback(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetContent(200, "fallback")
	return response, nil
}

func (u *QueryHandler) GetHandle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetContent(200, "OK")
	response.SetHeader("kkk", "mmmm")
	response.SetJWTBody(map[string]string{
		"id": "1",
	})
	return response, nil
}
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetStandardResponse()
	sharding_index, sharding_count := ctx.Request.Ext.GetSharding()
	ctx.Log.Infof("sharding:index-%d,count-%d", sharding_index, sharding_count)
	response.SetContent(200, "success")
	ctx.Log.Info("form:id:", ctx.Request.Form.GetString("id"))
	response.SetJWTBody(map[string]string{
		"id": "1",
	})
	return response, nil
}

func (u *QueryHandler) Close() error {
	return nil
}
