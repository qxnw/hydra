package registry

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/hydra/context"
)

func (s *registryProxy) updateValue(ctx *context.Context) (r string, err error) {
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
	if !b {
		err = fmt.Errorf("节点不存在:%s", path)
		return
	}
	value, err := input.Get("value")
	if err != nil {
		err = fmt.Errorf("缺少输入参数value")
		return
	}
	version, err := input.Get("version")
	if err != nil {
		err = fmt.Errorf("缺少输入参数version")
		return
	}
	v, err := strconv.Atoi(version)
	if err != nil {
		err = fmt.Errorf("输入参数version不是有效的数字")
		return
	}
	err = s.registry.Update(path, value, int32(v))
	if err == nil {
		return "SUCCESS", nil
	}
	_, ov, err1 := s.registry.GetValue(path)
	if err1 != nil {
		return "", err
	}
	if ov != int32(v) {
		err = fmt.Errorf("更新数据的版本错误，已发现最新版本:%d", ov)
	}
	return "", err
}
