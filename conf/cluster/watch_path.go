package cluster

import (
	"fmt"
	"sync"
	"time"

	"errors"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
)

type watchPath struct {
	updater      chan *conf.Updater
	cacheAddress cmap.ConcurrentMap
	exists       bool
	path         string
	registry     registry.Registry
	timeSpan     time.Duration
	closeChan    chan struct{}
	serverName   string
	domain       string
	done         bool
	mu           sync.Mutex
	*logger.Logger
}

func NewWatchPath(domain string, serverName string, path string, registry registry.Registry, updater chan *conf.Updater, timeSpan time.Duration, log *logger.Logger) *watchPath {
	return &watchPath{
		domain:       domain,
		serverName:   serverName,
		registry:     registry,
		updater:      updater,
		timeSpan:     timeSpan,
		path:         path,
		cacheAddress: cmap.New(2),
		Logger:       log,
		closeChan:    make(chan struct{}),
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
		case <-w.closeChan:
			return errors.New("watch is closing")
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
	close(w.closeChan)
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchConf).Close()
		return true
	})
}
func (w *watchPath) checkChildrenChange(children []string, version int32) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, v := range children { //检查当前配置地址未缓存
		for _, sv := range conf.WatchServices { //hydra/servers/merchant.api/rpc/conf/conf
			name := fmt.Sprintf("%s/%s/%s/conf/%s", w.path, v, sv, w.serverName)
			if _, ok := w.cacheAddress.Get(name); !ok {
				w.cacheAddress.SetIfAbsentCb(name, func(input ...interface{}) (interface{}, error) {
					path := input[0].(string)
					f := NewWatchConf(w.domain, v, sv, path, w.registry, w.updater, w.timeSpan, w.Logger)
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
	START:
		for _, v := range children {
			for _, sv := range conf.WatchServices {
				exists = key == fmt.Sprintf("%s/%s/%s/conf/%s", w.path, v, sv, w.serverName)
				if exists {
					break START
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
