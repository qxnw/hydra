package register

import (
	"fmt"

	"strings"

	"time"

	"github.com/qxnw/hydra/registry"
)

type clusterServiceRegister struct {
	done          bool
	registry      registry.Registry
	crossRegistry registry.Registry

	serverName            string
	domain                string
	localTmpServices      map[string]string
	localSeqServices      map[string]string
	crossTmpServices      map[string]string
	crossSeqServices      map[string]string
	localRalationService  map[string]string
	crosslRalationService map[string]string

	closer chan struct{}
}

//newClusterServiceRegister 创建zookeeper配置文件监控器
func newClusterServiceRegister(domain string, serverName string, r registry.Registry, cross registry.Registry) (w *clusterServiceRegister) {
	w = &clusterServiceRegister{
		registry:              r,
		crossRegistry:         cross,
		domain:                domain,
		serverName:            serverName,
		localTmpServices:      make(map[string]string),
		localSeqServices:      make(map[string]string),
		crossTmpServices:      make(map[string]string),
		crossSeqServices:      make(map[string]string),
		localRalationService:  make(map[string]string),
		crosslRalationService: make(map[string]string),
		closer:                make(chan struct{}),
	}
	go w.checker()
	return
}

//Register 服务注册
func (w *clusterServiceRegister) RegisterTempNode(serviceName string, endPointName string, data string) (r string, err error) {
	path := fmt.Sprintf("/%s/services/%s/%s/providers/%s", strings.Trim(w.domain, "/"), w.serverName, strings.Trim(serviceName, "/"), endPointName)
	w.crossRegister(path, data)
	w.Unregister(path)
	err = w.registry.CreateTempNode(path, data)
	if err != nil {
		err = fmt.Errorf("service.registry.failed:%s(err:%v)", path, err)
		return
	}
	if b, err := w.registry.Exists(path); !b || err != nil {
		err = fmt.Errorf("service.registry.failed:节点不存在%s(err:%v)", path, err)
		return "", err
	}
	w.localTmpServices[path] = data
	return path, err
}
func (w *clusterServiceRegister) RegisterSeqNode(path string, data string) (r string, err error) {
	path = path + "_"
	w.crossSeqRegister(path, data)
	r, err = w.registry.CreateSeqNode(path, data)
	if err != nil {
		err = fmt.Errorf("service.registry.failed:%s(err:%v)", path, err)
		return
	}
	if b, err := w.registry.Exists(r); !b || err != nil {
		err = fmt.Errorf("service.registry.failed:节点不存在%s(err:%v)", r, err)
		return "", err
	}
	w.localSeqServices[path] = data
	w.localRalationService[r] = path
	return
}

//UnRegister 取消服务注册
func (w *clusterServiceRegister) Unregister(path string) error {
	//local.seq
	if r, ok := w.localRalationService[path]; ok {
		err := w.registry.Delete(path)
		if err != nil {
			return err
		}
		delete(w.localRalationService, path)
		delete(w.localSeqServices, r)
		w.crossDelete(r)
		return nil
	}
	//local.temp
	err := w.registry.Delete(path)
	if err != nil {
		return err
	}
	delete(w.localTmpServices, path)
	w.crossDelete(path)
	return nil
}
func (w *clusterServiceRegister) crossRegister(path, data string) {
	if w.crossRegistry != nil {
		w.crossRegistry.CreateTempNode(path, data)
		w.crossTmpServices[path] = data
	}
}
func (w *clusterServiceRegister) crossSeqRegister(path, data string) {
	if w.crossRegistry != nil {
		r, err := w.crossRegistry.CreateSeqNode(path, data)
		if err != nil {
			return
		}
		w.crossSeqServices[path] = data
		w.crosslRalationService[path] = r
	}
}
func (w *clusterServiceRegister) crossDelete(path string) {
	if w.crossRegistry != nil {
		if r, ok := w.crosslRalationService[path]; ok {
			w.crossRegistry.Delete(r)
			delete(w.crossSeqServices, path)
			delete(w.crosslRalationService, path)
			return
		}
		w.crossRegistry.Delete(path)
		delete(w.crossTmpServices, path)
	}
}
func (w *clusterServiceRegister) checker() {
LOOP:
	for {
		select {
		case <-w.closer:
			break LOOP
		case <-time.After(time.Second * 60):
			if w.registry == nil {
				continue
			}
			for k, v := range w.localTmpServices {
				if b, err := w.registry.Exists(k); err == nil && !b {
					err := w.registry.CreateTempNode(k, v)
					if err != nil {
						goto END
					}
				}
			}
			for k, v := range w.localRalationService {
				if b, err := w.registry.Exists(k); err == nil && !b {
					r, err := w.registry.CreateSeqNode(v, w.localSeqServices[v])
					if err != nil {
						goto END
					}
					delete(w.localRalationService, k)
					w.localRalationService[r] = v
				}
			}
			if w.crossRegistry == nil {
				continue
			}
			for k, v := range w.crossTmpServices {
				if b, err := w.crossRegistry.Exists(k); err == nil && !b {
					err := w.crossRegistry.CreateTempNode(k, v)
					if err != nil {
						goto END
					}
				}
			}
			for k, v := range w.crosslRalationService {
				if b, err := w.crossRegistry.Exists(v); err == nil && !b {
					r, err := w.crossRegistry.CreateSeqNode(k, w.crossSeqServices[k])
					if err != nil {
						goto END
					}
					delete(w.crosslRalationService, k)
					w.localRalationService[k] = r
				}
			}
		END:
		}
	}
}

//Close 关闭所有监控项
func (w *clusterServiceRegister) Close() error {
	close(w.closer)
	w.registry.Close()
	return nil
}
