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

func (s *registryProxy) getChildren(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	p, err := ctx.Input.Get("path")
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
	response.Success(string(buff))
	return
}
