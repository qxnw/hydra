package registry

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *registryProxy) createSEQPath(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	path, err := ctx.Input.Get("path")
	if err != nil {
		err = fmt.Errorf("缺少输入参数path")
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
	value, err := ctx.Input.Get("value")
	if err != nil {
		err = fmt.Errorf("缺少输入参数value")
		return
	}
	p, err := s.registry.CreateSeqNode(path, value)
	if err != nil {
		return
	}
	response.Success(p)
	return
}
