package registry

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/context"

	"strings"
)

type nodeValue struct {
	Content json.RawMessage `json:"content"`
	Version int32           `json:"version"`
}

//获取指定path的值
func (s *registryProxy) getValue(ctx *context.Context) (r string, st int, err error) {
	p, err := ctx.GetInput().Get("path")
	if err != nil {
		err = fmt.Errorf("缺少输入参数path")
		return
	}
	buffer, v, err := s.registry.GetValue(p)
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
	r = string(buff)
	return
}
