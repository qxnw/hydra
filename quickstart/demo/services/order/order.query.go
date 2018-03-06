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
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetContent(200, "hello world")
	_, err = u.container.GetDB("db")
	if err != nil {
		response.SetContent(0, err)
	}
	return response, nil
}

func (u *QueryHandler) Close() error {
	return nil
}
