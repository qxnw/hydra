package service

import (
	"fmt"

	"strings"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/utility"
)

type standaloneServiceRegister struct {
	done bool
	registry.Registry
	serverName string
	domain     string
}

//newClusterServiceRegister 创建zookeeper配置文件监控器
func newStandaloneServiceRegister(domain string, serverName string, r registry.Registry) (w *standaloneServiceRegister) {
	return &standaloneServiceRegister{
		Registry:   r,
		domain:     domain,
		serverName: serverName,
	}
}

//Register 服务注册
func (w *standaloneServiceRegister) RegisterService(serviceName string, endPointName string, data string) (string, error) {
	path := fmt.Sprintf("/%s/services/rpc/%s/%s/providers/%s", strings.Trim(w.domain, "/"), w.serverName, strings.Trim(serviceName, "/"), endPointName)
	return path, w.Registry.CreateTempNode(path, data)
}
func (w *standaloneServiceRegister) RegisterSeqNode(path string, data string) (string, error) {
	rp := path + "_" + utility.GetGUID()
	return rp, w.Registry.CreateTempNode(rp, data)
}

//UnRegister 取消服务注册
func (w *standaloneServiceRegister) Unregister(path string) error {
	return w.Registry.Delete(path)
}

//Close 关闭所有监控项
func (w *standaloneServiceRegister) Close() error {
	return nil
}

//standaloneResolver 注册中心解析器
type standaloneResolver struct {
}

//Resolve 从服务器获取数据
func (j *standaloneResolver) Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (c IService, err error) {

	r, err := registry.NewLocalRegistry()
	if err != nil {
		return
	}
	c = newStandaloneServiceRegister(domain, serverName, r)
	return
}

func init() {
	Register("standalone", &standaloneResolver{})
}
