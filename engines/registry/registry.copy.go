package registry

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
)

type kv struct {
	path  string
	vp    string
	value []byte
}

//Copy 备份注册中心所有节点
func Copy(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		domain, err := ctx.Request.Form.Get("fromDomain")
		if err != nil {
			err = fmt.Errorf("缺少输入参数fromDomain")
			return
		}

		ndomain, err := ctx.Request.Form.Get("toDomain")
		if err != nil {
			err = fmt.Errorf("缺少输入参数toDomain")
			return
		}
		registry := c.GetRegistry()
		serverData, err := getCopyNodes(registry, fmt.Sprintf("/%s/servers", domain), "servers")
		if err != nil {
			return
		}

		varData, err := getCopyNodes(registry, fmt.Sprintf("/%s/var", domain), "var")
		if err != nil {
			return
		}
		for _, v := range serverData {
			path := fmt.Sprintf("/%s/%s", ndomain, v.vp)
			err = registry.CreatePersistentNode(path, string(v.value))
			if err != nil {
				return
			}
		}
		for _, v := range varData {
			path := fmt.Sprintf("/%s/%s", ndomain, v.vp)
			err = registry.CreatePersistentNode(path, string(v.value))
			if err != nil {
				return
			}
		}
		response.Success()
		return
	}
}
func getCopyNodes(r registry.Registry, p string, ch string) (re []kv, err error) {
	re = make([]kv, 0, 2)
	data, _, _ := r.GetValue(p)
	if err != nil {
		err = fmt.Errorf("未找到节点数据:%s(err:%v)", p, err)
		return
	}
	if len(data) > 0 {
		re = append(re, kv{path: p, vp: ch, value: data})
	}
	children, _, err := r.GetChildren(p)
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
		cd, err := getCopyNodes(r, fmt.Sprintf("%s/%s", p, v), fmt.Sprintf("%s/%s", ch, v))
		if err != nil {
			return nil, err
		}
		re = append(re, cd...)
	}
	return re, nil
}
