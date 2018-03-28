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
func (u *QueryHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	db, err := u.container.GetDB("db")
	if err != nil {
		return err
	}
	row, _, _, err := db.Query("select * from sys_system_dictionary", map[string]interface{}{})
	if err != nil {
		return err
	}
	return row
}
