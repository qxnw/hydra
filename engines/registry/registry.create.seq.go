package registry

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//CreateSEQNode 创建序列节点
func CreateSEQNode(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
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
		p, err := registry.CreateSeqNode(path, value)
		if err != nil {
			return
		}
		response.Success(p)
		return
	}
}
