package activity

import (
	"github.com/qxnw/collector/modules/activity"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

type ActivityHandler struct {
	fields    map[string][]string
	container component.IContainer
	act       activity.IActivity
}

func NewActivityHandler(container component.IContainer) (u *ActivityHandler, err error) {
	u = &ActivityHandler{
		container: container,
		fields: map[string][]string{
			"input": []string{"id"},
		},
	}
	u.act, err = activity.GetActivity(container)
	if err != nil {
		return
	}
	return
}
func (u *ActivityHandler) Fallback(name string, mode string, service string, ctx *context.Context) (response *context.ObjectResponse, err error) {
	response = context.GetObjectResponse()
	db, err := u.container.GetDefaultDB()
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

func (u *ActivityHandler) Handle(name string, mode string, service string, ctx *context.Context) (response *context.ObjectResponse, err error) {
	response = context.GetObjectResponse()
	db, err := u.container.GetDefaultDB()
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

func (u *ActivityHandler) Close() error {
	return nil
}
