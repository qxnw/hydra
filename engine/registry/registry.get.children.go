package registry

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/context"
)

type kv struct {
	path  string
	vp    string
	value []byte
}

func (s *registryProxy) getChildren(ctx *context.Context) (r string, st int, err error) {
	p, err := ctx.GetInput().Get("path")
	if err != nil {
		err = fmt.Errorf("缺少输入参数path")
		return
	}
	children, version, err := s.registry.GetChildren(p)
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
	r = string(buff)
	return
}
