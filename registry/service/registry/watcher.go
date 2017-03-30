package registry

import (
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf/service"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type serviceWatcher struct {
	watchRootChan  chan string
	watchPaths     cmap.ConcurrentMap
	notifyConfChan chan []*service.ServiceUpdater
	isInitialized  bool
	done           bool
	registry       Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	tag            string
}

//NewServiceWatcher 创建zookeeper配置文件监控器
func NewServiceWatcher(domain string, tag string, registry Registry) (w *serviceWatcher) {
	w = &serviceWatcher{
		watchRootChan:  make(chan string, 10),
		notifyConfChan: make(chan []*service.ServiceUpdater, 1),
		watchPaths:     cmap.New(),
		registry:       registry,
		timeSpan:       time.Second,
		domain:         domain,
		tag:            tag,
	}
	return
}

//Start 启用配置文件监控
func (w *serviceWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	path := fmt.Sprintf("%s/services/%s", w.domain, w.tag)
	w.watchRootChan <- path

	go w.watch()
	return nil
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
			watcher := NewWatchPath(w.domain, w.tag, p, w.registry, w.notifyConfChan, w.timeSpan)
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

//Notify 节点变化后通知
func (w *serviceWatcher) Notify() chan []*service.ServiceUpdater {
	return w.notifyConfChan
}
