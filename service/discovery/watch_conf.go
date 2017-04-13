package discovery

import (
	"errors"
	"time"

	"sync"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type watchConf struct {
	path           string
	registry       registry.Registry
	timeSpan       time.Duration
	notifyConfChan chan []*registry.ServiceUpdater
	cacheAddress   cmap.ConcurrentMap
	done           bool
	notifyCount    int
	mu             sync.Mutex
	exists         bool
}

func NewWatchConf(path string, registry registry.Registry, updater chan []*registry.ServiceUpdater, timeSpan time.Duration) *watchConf {
	return &watchConf{path: path,
		registry:       registry,
		notifyConfChan: updater,
		cacheAddress:   cmap.New(),
		timeSpan:       timeSpan}
}

//watchConf 监控配置项变化，当发生错误时持续监控节点变化，只有明确节点不存在时才会通知关闭
func (w *watchConf) watch() (err error) {
LOOP:
	w.exists = false
	isExists, _ := w.registry.Exists(w.path)
	for !isExists {
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				return errors.New("watcher is closing")
			}
			if isExists, err = w.registry.Exists(w.path); !isExists && err == nil {
				w.NotifyConfDel()
			}
		}
	}
	children, version, err := w.registry.GetChildren(w.path)
	if err != nil {
		goto LOOP
	}
	w.exists = isExists
	w.checkChildrenChange(children, version)
	//监控子节点变化
	ch, err := w.registry.WatchChildren(w.path)
	if err != nil {
		goto LOOP
	}

	for {
		select {
		case children, ok := <-ch:
			if !ok || w.done {
				return errors.New("watch is closing")
			}
			if err = children.GetError(); err != nil {
				goto LOOP
			}
			w.checkChildrenChange(children.GetValue())
			//继续监控子节点变化
			ch, err = w.registry.WatchChildren(w.path)
			if err != nil {
				goto LOOP
			}
		}
	}
}

func (w *watchConf) NotifyConfDel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		updater := &registry.ServiceUpdater{Value: key, Op: registry.DEL}
		w.notifyConfChan <- []*registry.ServiceUpdater{updater}
		return true
	})
	w.exists = false
}
func (w *watchConf) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true

}

func (w *watchConf) checkChildrenChange(children []string, version int32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	updaters := make([]*registry.ServiceUpdater, 0, 0)
	w.cacheAddress.IterCb(func(key string, value interface{}) bool {
		return true
	})
	for _, v := range children { //检查当前配置地址未缓存
		if _, ok := w.cacheAddress.Get(v); !ok {
			if ok, _, _ := w.cacheAddress.SetIfAbsentCb(v, func(input ...interface{}) (interface{}, error) {
				name := input[0].(string)
				return name, nil
			}, v); ok {
				updaters = append(updaters, &registry.ServiceUpdater{Op: registry.ADD, Value: v})
			}
		}

	}
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		exists := false
		for _, v := range children {
			exists = key == v
			if exists {
				break
			}

		}
		if !exists {
			updaters = append(updaters, &registry.ServiceUpdater{Op: registry.DEL, Value: key})
			return true
		}
		return false
	})
	if len(updaters) > 0 {
		w.notifyConfChan <- updaters
	}

}
