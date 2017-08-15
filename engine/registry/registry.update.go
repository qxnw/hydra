package registry

import (
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/context"
)

//根据path,value,version更新指定节点的值
func (s *registryProxy) updateValue(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
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
	if !b {
		err = fmt.Errorf("节点不存在:%s", path)
		return
	}
	value, err := ctx.Input.Get("value")
	if err != nil {
		err = fmt.Errorf("缺少输入参数value")
		return
	}
	version, err := ctx.Input.Get("version")
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
		response.Success()
		return
	}
	_, ov, err1 := s.registry.GetValue(path)
	if err1 != nil {
		return
	}
	if ov != int32(v) {
		err = fmt.Errorf("更新数据的版本错误，已发现最新版本:%d", ov)
		response.SetStatus(409)
	}
	return
}
