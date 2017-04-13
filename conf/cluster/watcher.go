package cluster

import (
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type registryConfWatcher struct {
	watchRootChan  chan string
	watchPaths     cmap.ConcurrentMap
	notifyConfChan chan *conf.Updater
	isInitialized  bool
	done           bool
	registry       registry.Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	tag            string
}

//NewRegistryConfWatcher 创建zookeeper配置文件监控器
func NewRegistryConfWatcher(domain string, tag string, registry registry.Registry) (w *registryConfWatcher) {
	w = &registryConfWatcher{
		watchRootChan:  make(chan string, 10),
		notifyConfChan: make(chan *conf.Updater, 1),
		watchPaths:     cmap.New(),
		registry:       registry,
		timeSpan:       time.Second,
		domain:         domain,
		tag:            tag,
	}
	return
}

//Start 启用配置文件监控
func (w *registryConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	path := fmt.Sprintf("%s/servers", w.domain)
	w.watchRootChan <- path

	go w.watch()
	return nil
}

//watch 监控配置路径变化和配置数据变化
func (w *registryConfWatcher) watch() {
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
func (w *registryConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.watchPaths.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchPath).Close()
		return true
	})
	return nil
}

//Notify 节点变化后通知
func (w *registryConfWatcher) Notify() chan *conf.Updater {
	return w.notifyConfChan
}
