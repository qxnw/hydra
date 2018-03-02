package order

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type OrderHandler struct {
	fields    map[string][]string
	container component.IContainer
}

func NewOrderHandler(container component.IContainer) (u *OrderHandler, err error) {
	u = &OrderHandler{
		container: container,
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	return
}
func (u *OrderHandler) Fallback(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
	response = context.GetObjectResponse()
	db, err := u.container.GetDefaultDB()
	if err != nil {
		return
	}
	data, _, _, err := db.Query("select * from sys_dictionary_Info where rownum<=1", map[string]interface{}{})
	if err != nil {
		return
	}
	response.SetContent(200, data)
	return
}

func (u *OrderHandler) Handle(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
	response = context.GetObjectResponse()
	db, err := u.container.GetDefaultDB()
	if err != nil {
		return
	}
	data, _, _, err := db.Query("select * from sys_dictionary_Info where rownum<=1", map[string]interface{}{})
	if err != nil {
		return
	}
	response.SetContent(200, data)
	return
}

func (u *OrderHandler) Close() error {
	return nil
}
