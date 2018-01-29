package product

import (
	"github.com/qxnw/hydra/context"
)

type ProductHandler struct {
	fields map[string][]string
}

func NewProductHandler() (u *ProductHandler) {
	u = &ProductHandler{
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	return
}

func (u *ProductHandler) Handle(name string, mode string, service string, ctx *context.Context) (response *context.ObjectResponse, err error) {
	response = context.GetObjectResponse()
	response.Success("success")
	return
}

func (u *ProductHandler) Close() error {
	return nil
}
