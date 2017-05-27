package registry

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/context"

	"strings"

	"github.com/qxnw/lib4go/transform"
)

type nodeValue struct {
	Content json.RawMessage `json:"content"`
	Version int32           `json:"version"`
}

func (s *registryProxy) getValue(ctx *context.Context) (r string, err error) {
	input, err := s.getGetParams(ctx)
	if err != nil {
		return
	}
	p, err := input.Get("path")
	if err != nil {
		err = fmt.Errorf("缺少输入参数path")
		return
	}
	if !strings.Contains(p, s.domain) {
		err = fmt.Errorf("path路径必须是:%s开头", s.domain)
		return
	}
	buffer, v, err := s.registry.GetValue(p)
	if err != nil {
		return
	}
	obj := &nodeValue{Content: buffer, Version: v}
	buff, err := json.Marshal(obj)
	if err != nil {
		return
	}
	r = string(buff)
	return
}

func (s *registryProxy) getGetParams(ctx *context.Context) (input transform.ITransformGetter, err error) {
	if ctx.Input.Input == nil {
		err = fmt.Errorf("input不能为空:%v", ctx.Input)
		return
	}
	input = ctx.Input.Input.(transform.ITransformGetter)
	return
}
