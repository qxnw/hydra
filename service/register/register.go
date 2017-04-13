package register

import (
	"fmt"

	"github.com/qxnw/hydra/registry"
)

type serviceRegister struct {
	done     bool
	registry registry.Registry
	sysName  string
	domain   string
}

//NewRegister 创建zookeeper配置文件监控器
func newServiceRegister(domain string, sysName string, r registry.Registry) (w *serviceRegister) {
	return &serviceRegister{
		registry: r,
		domain:   domain,
		sysName:  sysName,
	}
}

//Register 服务注册
func (w *serviceRegister) Register(serviceName string, endPointName string, data string) (string, error) {
	path := fmt.Sprintf("%s/services/%s/%s/providers/%s", w.domain, w.sysName, serviceName, endPointName)
	return path, w.registry.CreateTempNode(path, data)
}

//UnRegister 取消服务注册
func (w *serviceRegister) Unregister(path string) error {
	return w.registry.Delete(path)
}

//Close 关闭所有监控项
func (w *serviceRegister) Close() error {
	return nil
}
