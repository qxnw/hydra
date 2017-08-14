package registry

import (
	"fmt"
	"time"

	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/file"
)

//saveAll 无需任何输入参数，直接备份当前域所在目录下的所有配置
func (s *registryProxy) saveAll(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	serverData, err := s.getChildrenNodes(fmt.Sprintf("%s/servers", s.ctx.Domain))
	if err != nil {
		return
	}

	varData, err := s.getChildrenNodes(fmt.Sprintf("%s/var", s.ctx.Domain))
	if err != nil {
		return
	}
	serverData = append(serverData, varData...)
	savePath := make([]string, 0, len(serverData))
	root := fmt.Sprintf("./bak/registry[%s]/%s/", strings.Replace(s.ctx.Registry, "/", "-", -1), time.Now().Format("20060102150405"))
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
	response.Success(fmt.Sprintf("success.%d", len(savePath)))
	return
}
func (s *registryProxy) getChildrenNodes(p string) (r []kv, err error) {
	r = make([]kv, 0, 2)
	data, _, _ := s.registry.GetValue(p)
	if err != nil {
		err = fmt.Errorf("未找到节点数据:%s(err:%v)", p, err)
		return
	}
	if len(data) > 0 {
		r = append(r, kv{path: p, value: data})
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
		cd, err := s.getChildrenNodes(fmt.Sprintf("%s/%s", p, v))
		if err != nil {
			return nil, err
		}
		r = append(r, cd...)
	}
	return r, nil
}
