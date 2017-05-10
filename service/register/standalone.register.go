package register

import (
	"fmt"

	"strings"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/utility"
)

type standaloneServiceRegister struct {
	done       bool
	registry   registry.Checker
	serverName string
	domain     string
}

//newClusterServiceRegister 创建zookeeper配置文件监控器
func newStandaloneServiceRegister(domain string, serverName string, r registry.Checker) (w *standaloneServiceRegister) {
	return &standaloneServiceRegister{
		registry:   r,
		domain:     domain,
		serverName: serverName,
	}
}

//Register 服务注册
func (w *standaloneServiceRegister) Register(serviceName string, endPointName string, data string) (string, error) {
	path := fmt.Sprintf("/%s/services/%s/%s/providers/%s", strings.Trim(w.domain, "/"), w.serverName, strings.Trim(serviceName, "/"), endPointName)
	return path, w.registry.CreateFile(path, data)
}
func (w *standaloneServiceRegister) RegisterWithPath(path string, data string) (string, error) {
	rp := path + "_" + utility.GetGUID()
	return rp, w.registry.CreateFile(rp, data)
}

//UnRegister 取消服务注册
func (w *standaloneServiceRegister) Unregister(path string) error {
	return w.registry.Delete(path)
}

//Close 关闭所有监控项
func (w *standaloneServiceRegister) Close() error {
	return nil
}
