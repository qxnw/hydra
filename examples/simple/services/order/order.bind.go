package order

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type Input struct {
	ID   string `form:"id" json:"id" binding:"required"`
	Name string `form:"name" json:"name"`
}
type BindHandler struct {
	container component.IContainer
}

func NewBindHandler(container component.IContainer) (u *BindHandler) {

	return &BindHandler{container: container}
}
func (u *BindHandler) Handle(name string, engine string, service string, ctx *context.Context) (r context.Response, err error) {
	response := context.GetObjectResponse()
	var input Input
	if err := ctx.Request.Bind(&input); err != nil {
		response.SetContent(0, err)
		return response, err
	}
	ctx.Log.Infof("id:%s,%s", ctx.Request.Form.GetString("id"), ctx.Request.Form.GetString("name"))
	response.SetContent(200, input)
	return response, nil
}

func (u *BindHandler) Close() error {
	return nil
}
