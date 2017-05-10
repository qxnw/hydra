package register

import (
	"fmt"

	"strings"

	"github.com/qxnw/hydra/registry"
)

type clusterServiceRegister struct {
	done          bool
	registry      registry.Registry
	crossRegistry registry.Registry
	serverName    string
	domain        string
}

//newClusterServiceRegister 创建zookeeper配置文件监控器
func newClusterServiceRegister(domain string, serverName string, r registry.Registry, cross registry.Registry) (w *clusterServiceRegister) {
	return &clusterServiceRegister{
		registry:      r,
		crossRegistry: cross,
		domain:        domain,
		serverName:    serverName,
	}
}

//Register 服务注册
func (w *clusterServiceRegister) Register(serviceName string, endPointName string, data string) (r string, err error) {
	path := fmt.Sprintf("/%s/services/%s/%s/providers/%s", strings.Trim(w.domain, "/"), w.serverName, strings.Trim(serviceName, "/"), endPointName)
	w.crossRegister(path, data)
	w.Unregister(path)
	err = w.registry.CreateTempNode(path, data)
	if err != nil {
		err = fmt.Errorf("service.registry.failed:%s(err:%v)", path, err)
		return
	}
	return path, err
}
func (w *clusterServiceRegister) RegisterWithPath(path string, data string) (r string, err error) {
	rpath := path + "_"
	w.crossSeqRegister(rpath, data)
	r, err = w.registry.CreateSeqNode(rpath, data)
	if err != nil {
		err = fmt.Errorf("service.registry.failed:%s(err:%v)", rpath, err)
		return
	}
	return
}

//UnRegister 取消服务注册
func (w *clusterServiceRegister) Unregister(path string) error {
	w.crossDelete(path)
	return w.registry.Delete(path)
}
func (w *clusterServiceRegister) crossRegister(path, data string) {
	if w.crossRegistry != nil {
		w.crossRegistry.CreateTempNode(path, data)
	}
}
func (w *clusterServiceRegister) crossSeqRegister(path, data string) {
	if w.crossRegistry != nil {
		w.crossRegistry.CreateSeqNode(path, data)
	}
}
func (w *clusterServiceRegister) crossDelete(path string) {
	if w.crossRegistry != nil {
		w.crossRegistry.Delete(path)
	}
}

//Close 关闭所有监控项
func (w *clusterServiceRegister) Close() error {
	w.registry.Close()
	return nil
}
