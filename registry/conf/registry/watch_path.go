package registry

import (
	"fmt"
	"sync"
	"time"

	"errors"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/registry/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type watchPath struct {
	updater       chan *conf.Updater
	cacheAddress  cmap.ConcurrentMap
	exists        bool
	watchRootChan chan string
	path          string
	registry      registry.Registry
	timeSpan      time.Duration
	tag           string
	domain        string
	done          bool
	mu            sync.Mutex
}

func NewWatchPath(domain string, tag string, path string, registry registry.Registry, updater chan *conf.Updater, timeSpan time.Duration) *watchPath {
	return &watchPath{
		cacheAddress:  cmap.New(),
		watchRootChan: make(chan string, 1),
		domain:        domain,
		tag:           tag,
		registry:      registry,
		timeSpan:      timeSpan,
		path:          path,
		updater:       updater,
	}

}

//watchPath 监控当前节点是否存在，不存在时也持续监控只到当前监控被关闭
//节点存在时，获取所有子节点，并启动配置路径监控
//节点由存在变为不存在时，关闭所有子节点
func (w *watchPath) watch() (err error) {
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
				w.notifyPathDel()
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

//notifyPathDel 关闭所有配置项的监控
func (w *watchPath) notifyPathDel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.exists { //目录已删除
		w.exists = false
		w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
			value.(*watchConf).NotifyConfDel()
			return true
		})
	}
}

//Close 推出当前流程，并闭所有子流程
func (w *watchPath) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchConf).Close()
		return true
	})
}
func (w *watchPath) checkChildrenChange(children []string, version int32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, v := range children { //检查当前配置地址未缓存
		for _, sv := range conf.WatchServices {
			name := fmt.Sprintf("%s/%s/%s/conf/%s", w.path, v, sv, w.tag)
			if _, ok := w.cacheAddress.Get(name); !ok {
				w.cacheAddress.SetIfAbsentCb(name, func(input ...interface{}) (interface{}, error) {
					path := input[0].(string)
					f := NewWatchConf(path, w.registry, w.updater, w.timeSpan)
					f.args = map[string]string{
						"domain": w.domain,
						"root":   fmt.Sprintf("%s/%s/%s/conf", w.path, v, sv),
						"path":   name,
					}
					go f.watch()
					return f, nil
				}, name)
			}
		}

	}
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		exists := false
		for _, v := range children {
			for _, sv := range conf.WatchServices {
				exists = key == fmt.Sprintf("%s/%s/%s/conf/%s", w.path, v, sv, w.tag)
				if exists {
					break
				}
			}
		}
		if !exists {
			value.(*watchConf).NotifyConfDel()
			value.(*watchConf).Close()
			return true
		}
		return false
	})
}