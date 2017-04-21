package register

import (
	"fmt"

	"strings"

	"github.com/qxnw/hydra/registry"
)

type serviceRegister struct {
	done       bool
	registry   registry.Registry
	serverName string
	domain     string
}

//NewRegister 创建zookeeper配置文件监控器
func newServiceRegister(domain string, serverName string, r registry.Registry) (w *serviceRegister) {
	return &serviceRegister{
		registry:   r,
		domain:     domain,
		serverName: serverName,
	}
}

//Register 服务注册
func (w *serviceRegister) Register(serviceName string, endPointName string, data string) (string, error) {
	path := fmt.Sprintf("/%s/services/%s/%s/providers/%s", strings.Trim(w.domain, "/"), w.serverName, strings.Trim(serviceName, "/"), endPointName)
	return path, w.registry.CreateTempNode(path, data)
}
func (w *serviceRegister) RegisterWithPath(path string, data string) (string, error) {
	return path, w.registry.CreateTempNode(path, data)
}

//UnRegister 取消服务注册
func (w *serviceRegister) Unregister(path string) error {
	return w.registry.Delete(path)
}

//Close 关闭所有监控项
func (w *serviceRegister) Close() error {
	w.registry.Close()
	return nil
}
