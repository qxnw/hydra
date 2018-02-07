package product

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type ProductHandler struct {
	fields    map[string][]string
	container component.IContainer
}

func NewProductHandler(container component.IContainer) (u *ProductHandler) {
	u = &ProductHandler{
		container: container,
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	return
}

func (u *ProductHandler) Handle(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	queue, err := u.container.GetDefaultQueue()
	if err != nil {
		return
	}
	err = queue.Push("/hydra/:mqc/:t1", "100")
	if err != nil {
		return
	}
	response.Success(`success`)
	return
}

func (u *ProductHandler) Close() error {
	return nil
}
