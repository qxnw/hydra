package main

import "github.com/qxnw/hydra/context"

type OrderHandler struct {
	fields map[string][]string
}

func NewOrderHandler() (u *OrderHandler, err error) {
	u = &OrderHandler{
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	return
}
func (u *OrderHandler) Fallback(name string, mode string, service string, ctx *context.Context) (response *context.ObjectResponse, err error) {
	response = context.GetObjectResponse()
	db, err := Container.GetDefaultDB()
	if err != nil {
		return
	}
	data, _, _, err := db.Query("select * from sys_dictionary_Info where rownum<=1", map[string]interface{}{})
	if err != nil {
		return
	}
	response.Success(data)
	return
}

func (u *OrderHandler) Handle(name string, mode string, service string, ctx *context.Context) (response *context.ObjectResponse, err error) {
	response = context.GetObjectResponse()
	db, err := Container.GetDefaultDB()
	if err != nil {
		return
	}
	data, _, _, err := db.Query("select * from sys_dictionary_Info where rownum<=1", map[string]interface{}{})
	if err != nil {
		return
	}
	response.Success(data)
	return
}

func (u *OrderHandler) Close() error {
	return nil
}
