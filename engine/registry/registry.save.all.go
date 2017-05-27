package registry

import (
	"fmt"
	"time"

	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/file"
)

type kv struct {
	path  string
	value []byte
}

func (s *registryProxy) saveAll(ctx *context.Context) (r string, err error) {
	serverData, err := s.getChildrenNodes(fmt.Sprintf("%s/servers", s.domain))
	if err != nil {
		return "", err
	}

	varData, err := s.getChildrenNodes(fmt.Sprintf("%s/var", s.domain))
	if err != nil {
		return "", err
	}
	serverData = append(serverData, varData...)
	savePath := make([]string, 0, len(serverData))
	root := fmt.Sprintf("./bak/registry[%s]/%s/", strings.Replace(s.registryAddrs, "/", "-", -1), time.Now().Format("20060102150405"))
	for _, v := range serverData {
		realPath := fmt.Sprintf("%s/%s", root, strings.Replace(v.path, "/", "-", -1))
		f, err := file.CreateFile(realPath)
		if err != nil {
			return "", err
		}
		_, err = f.Write(v.value)
		if err != nil {
			return "", err
		}
		err = f.Close()
		if err != nil {
			return "", err
		}
		savePath = append(savePath, realPath)
	}
	return fmt.Sprintf("success.%d", len(savePath)), nil
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
