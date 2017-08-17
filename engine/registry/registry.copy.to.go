package registry

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

//domainCopy 无需任何输入参数，直接备份当前域所在目录下的所有配置
func (s *registryProxy) domainCopy(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response =context.GetStandardResponse()
	domain, err := ctx.Input.Get("fromDomain")
	if err != nil {
		err = fmt.Errorf("缺少输入参数fromDomain")
		return
	}

	ndomain, err := ctx.Input.Get("toDomain")
	if err != nil {
		err = fmt.Errorf("缺少输入参数toDomain")
		return
	}

	serverData, err := s.getCopyNodes(fmt.Sprintf("%s/servers", domain), "servers")
	if err != nil {
		return
	}

	varData, err := s.getCopyNodes(fmt.Sprintf("%s/var", domain), "var")
	if err != nil {
		return
	}
	for _, v := range serverData {
		path := fmt.Sprintf("%s/%s", ndomain, v.vp)
		err = s.registry.CreatePersistentNode(path, string(v.value))
		if err != nil {
			return
		}
	}
	for _, v := range varData {
		path := fmt.Sprintf("%s/%s", ndomain, v.vp)
		err = s.registry.CreatePersistentNode(path, string(v.value))
		if err != nil {
			return
		}
	}
	response.Success()
	return
}
func (s *registryProxy) getCopyNodes(p string, ch string) (r []kv, err error) {
	r = make([]kv, 0, 2)
	data, _, _ := s.registry.GetValue(p)
	if err != nil {
		err = fmt.Errorf("未找到节点数据:%s(err:%v)", p, err)
		return
	}
	if len(data) > 0 {
		r = append(r, kv{path: p, vp: ch, value: data})
	}
	children, _, err := s.registry.GetChildren(p)
	if err != nil {
		err = fmt.Errorf("无法获取子节点数据:%s(err:%v)", p, err)
		return
	}
	if len(children) == 0 {
		return
	}
	for _, v := range children {
		if v == "servers" {
			continue
		}
		cd, err := s.getCopyNodes(fmt.Sprintf("%s/%s", p, v), fmt.Sprintf("%s/%s", ch, v))
		if err != nil {
			return nil, err
		}
		r = append(r, cd...)
	}
	return r, nil
}
