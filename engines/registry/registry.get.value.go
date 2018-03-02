package registry

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"strings"
)

type nodeValue struct {
	Content json.RawMessage `json:"content"`
	Version int32           `json:"version"`
}

//GetNodeValue 获取节点值
func GetNodeValue(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		p, err := ctx.Request.Form.Get("path")
		if err != nil {
			err = fmt.Errorf("缺少输入参数path")
			return
		}
		registry := c.GetRegistry()
		buffer, v, err := registry.GetValue(p)
		if err != nil {
			return
		}

		result := make(map[string]interface{})
		result["version"] = v
		content := string(buffer)
		if strings.Contains(content, "{") || strings.Contains(content, "[") {
			result["content"] = json.RawMessage(buffer)
		} else {
			result["content"] = content
		}
		buff, err := json.Marshal(result)
		if err != nil {
			return
		}
		response.SetContent(200, string(buff))
		return
	}
}
