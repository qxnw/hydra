package registry

import (
	"fmt"
	"time"

	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/file"
)

//Backup 备份注册中心所有节点
func Backup(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		registry := c.GetRegistry()
		serverData, err := getChildrenNodes(registry, fmt.Sprintf("/%s/servers", c.GetDomainName()))
		if err != nil {
			return
		}

		varData, err := getChildrenNodes(registry, fmt.Sprintf("/%s/var", c.GetDomainName()))
		if err != nil {
			return
		}
		serverData = append(serverData, varData...)
		savePath := make([]string, 0, len(serverData))
		root := fmt.Sprintf("./bak/registry[%s]/%s/", c.GetDomainName(), time.Now().Format("20060102150405"))
		for _, v := range serverData {
			realPath := fmt.Sprintf("%s/%s.json", root, strings.Replace(v.path, "/", "-", -1))
			f, err := file.CreateFile(realPath)
			if err != nil {
				return response, err
			}
			_, err = f.Write(v.value)
			if err != nil {
				return response, err
			}
			err = f.Close()
			if err != nil {
				return response, err
			}
			savePath = append(savePath, realPath)
		}
		response.SetContent(200, fmt.Sprintf("success.%d", len(savePath)))
		return
	}
}
func getChildrenNodes(rx registry.Registry, p string) (r []kv, err error) {
	r = make([]kv, 0, 2)
	data, _, _ := rx.GetValue(p)
	if err != nil {
		err = fmt.Errorf("未找到节点数据:%s(err:%v)", p, err)
		return
	}
	if len(data) > 0 {
		r = append(r, kv{path: p, value: data})
	}
	children, _, err := rx.GetChildren(p)
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
		cd, err := getChildrenNodes(rx, fmt.Sprintf("%s/%s", p, v))
		if err != nil {
			return nil, err
		}
		r = append(r, cd...)
	}
	return r, nil
}
