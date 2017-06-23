package discovery

import (
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type serviceDiscovery struct {
	watchRootChan  chan string
	watchPaths     cmap.ConcurrentMap
	notifyConfChan chan []*registry.ServiceUpdater
	isInitialized  bool
	done           bool
	registry       registry.Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	sysName        string
}

//newServiceDiscovery 创建zookeeper配置文件监控器
func newServiceDiscovery(domain string, sysName string, r registry.Registry) (w *serviceDiscovery) {
	w = &serviceDiscovery{
		watchRootChan:  make(chan string, 10),
		notifyConfChan: make(chan []*registry.ServiceUpdater, 1),
		watchPaths:     cmap.New(2),
		registry:       r,
		timeSpan:       time.Second,
		domain:         domain,
		sysName:        sysName,
	}
	return
}

//Notify 启用配置文件监控
func (w *serviceDiscovery) Notify() (chan []*registry.ServiceUpdater, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil, nil
	}
	w.isInitialized = true
	path := fmt.Sprintf("%s/services/%s", w.domain, w.sysName)
	w.watchRootChan <- path

	go w.watch()
	return w.notifyConfChan, nil
}

//Consume 服务消费订阅
func (w *serviceDiscovery) ConsumeCurrent(serviceName string, endPointName string, data string) error {
	return w.Consume(w.domain, w.sysName, serviceName, endPointName, data)
}

//Consume 服务消费订阅
func (w *serviceDiscovery) Consume(domain string, sysName string, serviceName string, endPointName string, data string) error {
	path := fmt.Sprintf("%s/services/%s/%s/consumers/%s_", domain, sysName, serviceName, endPointName)
	_, err := w.registry.CreateSeqNode(path, data)
	return err
}

//watch 监控配置路径变化和配置数据变化
func (w *serviceDiscovery) watch() {
START:
	for {
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				break START
			}
		case p, ok := <-w.watchRootChan:
			if w.done || !ok {
				break START
			}
			watcher := NewWatchPath(w.domain, w.sysName, p, w.registry, w.notifyConfChan, w.timeSpan)
			w.watchPaths.Set(p, watcher)
			go watcher.watch()
		}
	}
}

//Close 关闭所有监控项
func (w *serviceDiscovery) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.watchPaths.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchPath).Close()
		return true
	})
	return nil
}
