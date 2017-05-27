package registry

import (
	"fmt"

	"github.com/qxnw/hydra/context"

	"strings"
)

func (s *registryProxy) createPath(ctx *context.Context) (r string, err error) {
	input, err := s.getGetParams(ctx)
	if err != nil {
		return
	}
	path, err := input.Get("path")
	if err != nil {
		err = fmt.Errorf("缺少输入参数path")
		return
	}
	if !strings.Contains(path, s.domain) {
		err = fmt.Errorf("path路径必须是:%s开头", s.domain)
		return
	}
	b, err := s.registry.Exists(path)
	if err != nil {
		return
	}
	if b {
		err = fmt.Errorf("节点已经存在不能创建:%s", path)
		return
	}
	value, err := input.Get("value")
	if err != nil {
		err = fmt.Errorf("缺少输入参数value")
		return
	}
	err = s.registry.CreatePersistentNode(path, value)
	if err != nil {
		return
	}
	return "SUCCESS", nil
}
