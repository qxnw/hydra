package registry

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//GetChildrenNodes 获取所有子节点
func GetChildrenNodes(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		p, err := ctx.Request.Form.Get("path")
		if err != nil {
			err = fmt.Errorf("缺少输入参数path")
			return
		}
		registry := c.GetRegistry()
		children, version, err := registry.GetChildren(p)
		if err != nil {
			return
		}
		result := make(map[string]interface{})
		result["children"] = children
		result["version"] = version
		buff, err := json.Marshal(result)
		if err != nil {
			return
		}
		response.SetContent(200, string(buff))
		return
	}
}
