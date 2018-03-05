package main

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type QueryHandler struct {
}

func NewQueryHandler(container component.IContainer) (u *QueryHandler) {
	return &QueryHandler{}
}
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetStandardResponse()
	response.SetContent(200, "hello world")
	return response, nil
}

func (u *QueryHandler) Close() error {
	return nil
}
