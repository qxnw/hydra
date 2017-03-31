package service

import (
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type serviceWatcher struct {
	watchRootChan  chan string
	watchPaths     cmap.ConcurrentMap
	notifyConfChan chan []*ServiceUpdater
	isInitialized  bool
	done           bool
	registry       registry.Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	sysName        string
}

//NewServiceWatcher 创建zookeeper配置文件监控器
func NewServiceWatcher(domain string, sysName string, registry registry.Registry) (w *serviceWatcher) {
	w = &serviceWatcher{
		watchRootChan:  make(chan string, 10),
		notifyConfChan: make(chan []*ServiceUpdater, 1),
		watchPaths:     cmap.New(),
		registry:       registry,
		timeSpan:       time.Second,
		domain:         domain,
		sysName:        sysName,
	}
	return
}

//Notify 启用配置文件监控
func (w *serviceWatcher) Notify() (chan []*ServiceUpdater, error) {
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

//Publish 服务发布
func (w *serviceWatcher) Publish(serviceName string, endPointName string, data string) error {
	path := fmt.Sprintf("%s/services/%s/%s/providers/%s", w.domain, w.sysName, serviceName, endPointName)
	return w.registry.CreateTempNode(path, data)
}

//Consume 服务消费订阅
func (w *serviceWatcher) ConsumeCurrent(serviceName string, endPointName string, data string) error {
	return w.Consume(w.domain, w.sysName, serviceName, endPointName, data)
}

//Consume 服务消费订阅
func (w *serviceWatcher) Consume(domain string, sysName string, serviceName string, endPointName string, data string) error {
	path := fmt.Sprintf("%s/services/%s/%s/consumers/%s_", domain, sysName, serviceName, endPointName)
	_, err := w.registry.CreateSeqNode(path, data)
	return err
}

//watch 监控配置路径变化和配置数据变化
func (w *serviceWatcher) watch() {
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
func (w *serviceWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.watchPaths.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchPath).Close()
		return true
	})
	return nil
}
