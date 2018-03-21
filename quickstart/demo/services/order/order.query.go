package order

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type QueryHandler struct {
	container component.IContainer
	Name      string
}

func NewQueryHandler(container component.IContainer) (u *QueryHandler) {
	return &QueryHandler{container: container, Name: "QueryHandler"}
}
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetObjectResponse()
	response.SetContent(200, "success")
	return response, nil
}

func (u *QueryHandler) Close() error {
	return nil
}
