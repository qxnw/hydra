package order

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type Input struct {
	ID   string `form:"id" json:"id" valid:"required"` //绑定输入参数，并验证类型否是否必须输入
	Name string `form:"name" json:"name"`
}
type BindHandler struct {
	container component.IContainer
}

func NewBindHandler(container component.IContainer) (u *BindHandler) {
	return &BindHandler{container: container}
}
func (u *BindHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	var input Input
	if err := ctx.Request.Bind(&input); err != nil {
		return err
	}
	return input
}
