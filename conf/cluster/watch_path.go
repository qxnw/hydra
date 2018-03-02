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

type watchServer struct {
	updater      chan *conf.Updater
	cacheAddress cmap.ConcurrentMap
	exists       bool
	serverRoot   string
	registry     registry.Registry
	timeSpan     time.Duration
	closeChan    chan struct{}
	clientTag    string
	domain       string
	done         bool
	mu           sync.Mutex
	*logger.Logger
}

//newWatchServer 监控每个子系统的服务器节点变化
func newWatchServer(domain string, clientTag string, serverRoot string, registry registry.Registry, updater chan *conf.Updater, timeSpan time.Duration, log *logger.Logger) *watchServer {
	return &watchServer{
		domain:       domain,
		clientTag:    clientTag,
		registry:     registry,
		updater:      updater,
		timeSpan:     timeSpan,
		serverRoot:   serverRoot,
		cacheAddress: cmap.New(2),
		Logger:       log,
		closeChan:    make(chan struct{}),
	}

}

//watchPath 监控子系统、服务器变化
//服务器不存在时持续检查，直到子系统出现
func (w *watchServer) watch() (err error) {
LOOP:
	w.exists = false
	isExists, _ := w.registry.Exists(w.serverRoot)
	for !isExists { //检查服务器存储目录是否存在，不存在时持续进行检查
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				return errors.New("watcher is closing")
			}
			if isExists, err = w.registry.Exists(w.serverRoot); !isExists && err == nil {
				w.delWatchConf()
			}
		}
	}
	//获取系统列表，获取失败后持续进行检查
	children, version, err := w.registry.GetChildren(w.serverRoot)
	if err != nil {
		goto LOOP
	}
	w.exists = isExists

	//根据服务列表，查询并监控配置信息
	w.startWatchConf(children, version)

	//持续监控子系统变化
	ch, err := w.registry.WatchChildren(w.serverRoot)
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

			//根据服务列表，查询并监控配置信息
			w.startWatchConf(children.GetValue())
			ch, err = w.registry.WatchChildren(w.serverRoot)
			if err != nil {
				goto LOOP
			}
		}
	}
}

//delWatchConf 删除配置文件监控
func (w *watchServer) delWatchConf() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.exists { //目录已删除
		w.exists = false
		w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
			value.(*watchConf).notifyDeleted()
			return true
		})
	}
}

//Close 推出当前流程，并闭所有子流程
func (w *watchServer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	close(w.closeChan)
	w.cacheAddress.RemoveIterCb(func(key string, value interface{}) bool {
		value.(*watchConf).Close()
		return true
	})
}

//startWatchConf 启动服务器配置项监控
func (w *watchServer) startWatchConf(children []string, version int32) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, v := range children { //检查当前配置地址未缓存
		for _, sv := range conf.WatchServers { //hydra/servers/merchant.api/rpc/conf/conf
			//name := fmt.Sprintf("%s/%s/%s/conf/%s", w.serverRoot, v, sv, w.clientTag)
			//name := fmt.Sprintf("%s/%s/%s/%s", w.serverRoot, v, sv, w.clientTag)
			name := fmt.Sprintf("%s/%s/%s/%s/conf", w.serverRoot, v, sv, w.clientTag)
			if _, ok := w.cacheAddress.Get(name); !ok {
				w.cacheAddress.SetIfAbsentCb(name, func(input ...interface{}) (interface{}, error) {
					path := input[0].(string)
					f := newWatchConf(w.domain, v, sv, w.clientTag, path, w.registry, w.updater, w.timeSpan, w.Logger)
					f.args = map[string]string{
						"domain": w.domain,
						//"root":   fmt.Sprintf("%s/%s/%s/conf", w.serverRoot, v, sv),
						"root": fmt.Sprintf("%s/%s/%s/conf", w.serverRoot, v, sv),
						"path": name,
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
			for _, sv := range conf.WatchServers {
				//	exists = key == fmt.Sprintf("%s/%s/%s/conf/%s", w.serverRoot, v, sv, w.clientTag)
				exists = key == fmt.Sprintf("%s/%s/%s/%s/conf", w.serverRoot, v, sv, w.clientTag)
				if exists {
					break START
				}
			}
		}
		if !exists {
			value.(*watchConf).notifyDeleted()
			value.(*watchConf).Close()
			return true
		}
		return false
	})
}
