package cluster

import (
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
)

type registryConfWatcher struct {
	watchRootChan  chan string
	watchPaths     cmap.ConcurrentMap
	notifyConfChan chan *conf.Updater
	isInitialized  bool
	closeChan      chan struct{}
	done           bool
	registry       registry.Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	clientTag      string
	*logger.Logger
}

//newRegistryConfWatcher 创建基于注册中心的配置监控器
func newRegistryConfWatcher(domain string, clientTag string, registry registry.Registry, log *logger.Logger) (w *registryConfWatcher) {
	w = &registryConfWatcher{
		watchRootChan:  make(chan string, 1),
		notifyConfChan: make(chan *conf.Updater, 10),
		closeChan:      make(chan struct{}),
		watchPaths:     cmap.New(2),
		registry:       registry,
		timeSpan:       time.Second,
		domain:         strings.Trim(domain, "/"),
		clientTag:      clientTag,
		Logger:         log,
	}
	return
}

//Start 监控servers目录下服务器节点变化
func (w *registryConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	//path := fmt.Sprintf("/%s/conf", w.domain)
	path := fmt.Sprintf("/%s", w.domain)
	w.watchRootChan <- path

	go w.watch()
	return nil
}

//watch 启动实时监控，监控服务器节点变化
func (w *registryConfWatcher) watch() {
START:
	for {
		select {
		case <-w.closeChan:
			break START
		case path, ok := <-w.watchRootChan:
			if w.done || !ok {
				break START
			}
			w.watchPaths.SetIfAbsentCb(path, func(input ...interface{}) (interface{}, error) {
				p := input[0].(string)
				watcher := newWatchServer(w.domain, w.clientTag, p, w.registry, w.notifyConfChan, w.timeSpan, w.Logger)
				go watcher.watch()
				return watcher, nil
			}, path)

		}
	}
}

//Close 关闭所有监控项
func (w *registryConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.watchPaths.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchServer).Close()
		return true
	})
	w.registry.Close()
	close(w.closeChan)
	return nil
}

//Notify 节点变化后通知
func (w *registryConfWatcher) Notify() chan *conf.Updater {
	return w.notifyConfChan
}
