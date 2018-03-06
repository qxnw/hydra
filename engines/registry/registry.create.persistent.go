package registry

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//CreatePersistentNode 创建持续节点
func CreatePersistentNode(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		path, err := ctx.Request.Form.Get("path")
		if err != nil {
			err = fmt.Errorf("缺少输入参数path")
			return
		}
		registry := c.GetRegistry()
		b, err := registry.Exists(path)
		if err != nil {
			return
		}
		if b {
			err = fmt.Errorf("节点已经存在不能创建:%s", path)
			return
		}
		value, err := ctx.Request.Form.Get("value")
		if err != nil {
			err = fmt.Errorf("缺少输入参数value")
			return
		}
		err = registry.CreatePersistentNode(path, value)
		if err != nil {
			return
		}
		response.SetContent(200, "success")
		return
	}
}