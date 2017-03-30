package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strings"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/conf/server"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

type jsonConfWatcher struct {
	watchConfChan  chan string
	deleteConfChan chan string
	watchRootChan  chan string
	cacheAddress   cmap.ConcurrentMap
	cacheDir       cmap.ConcurrentMap
	notifyConfChan chan *server.Updater
	defTime        time.Time
	isInitialized  bool
	done           bool
	checker        checker
	timeSpan       time.Duration
	domain         string
	tag            string
	mu             sync.Mutex
}

type watcherPath struct {
	close   chan struct{}
	modTime time.Time
	conf    conf.Conf
	send    bool
}

//NewJSONConfWatcher 创建zookeeper配置文件监控器
func NewJSONConfWatcher(domain string, tag string) (w *jsonConfWatcher) {
	w = &jsonConfWatcher{
		notifyConfChan: make(chan *server.Updater),
		watchConfChan:  make(chan string, 2),
		deleteConfChan: make(chan string, 2),
		watchRootChan:  make(chan string, 10),
		cacheAddress:   cmap.New(),
		cacheDir:       cmap.New(),
		checker:        fileChecker{},
		domain:         domain,
		tag:            tag,
		timeSpan:       time.Second,
	}
	if tag == "" {
		w.tag = "conf"
	}
	w.defTime, _ = time.Parse("2000-01-01", "2006-01-02")
	return
}

//Start 启用配置文件监控
func (w *jsonConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	path := fmt.Sprintf("%s/servers", w.domain)
	w.cacheDir.Set(path, false)
	w.watchRootChan <- path

	go w.watch()
	return nil
}

//watch 监控配置路径变化和配置数据变化
func (w *jsonConfWatcher) watch() {
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
			go w.watchPath(p)
		case p, ok := <-w.watchConfChan:
			if w.done || !ok {
				break START
			}
			go w.watchConf(p)

		}
	}
}

//watchPath 监控配置路径变化, 文件夹不存在时判断是否已删除，已删除则关闭所有子目录监控，否则获取所有子目录并添加到监控目录
func (w *jsonConfWatcher) watchPath(path string) (err error) {
START:
	for {
		select {
		case del := <-w.deleteConfChan:
			w.cacheAddress.IterCb(func(key string, value interface{}) bool {
				if strings.HasPrefix(key, del) {
					value.(*watcherPath).close <- struct{}{}
				}
				return true
			})
		case <-time.After(w.timeSpan):
			if w.done {
				break START
			}
			b := w.checker.Exists(path)
			if v, ok := w.cacheDir.Get(path); ok && v.(bool) && !b { //目录已删除
				w.deleteConfChan <- path
				w.cacheDir.Set(path, false)
			}
			if !b {
				continue
			}
			children, err := w.checker.ReadDir(path)
			if err != nil {
				continue
			}
			w.cacheDir.Set(path, true)
			for _, v := range children { //检查当前配置地址未缓存
				for _, sv := range server.WatchServices {
					name := fmt.Sprintf("%s/%s/%s/conf/%s.json", path, v, sv, w.tag)
					if _, ok := w.cacheAddress.Get(name); !ok {
						w.cacheAddress.Set(name, &watcherPath{modTime: w.defTime, close: make(chan struct{}, 1)})
						w.watchConfChan <- name
					}
				}
			}
			w.cacheAddress.IterCb(func(key string, value interface{}) bool {
				exists := false
				for _, v := range children {
					for _, sv := range server.WatchServices {
						if key == fmt.Sprintf("%s/%s/%s/conf/%s.json", path, v, sv, w.tag) {
							exists = true
							break
						}
					}
				}
				if !exists {
					value.(*watcherPath).close <- struct{}{}
				}
				return false
			})
		}
	}
	return
}

//watchConf 监控配置项变化，当配置文件删除后关闭监控并发送删除通知，发生变化则只发送变更通知
func (w *jsonConfWatcher) watchConf(path string) (err error) {
	c, _ := w.cacheAddress.Get(path)
	ch := c.(*watcherPath).close
START:
	for {
		select {
		case <-ch:
			if w.done {
				break START
			}
			if c, ok := w.cacheAddress.Get(path); ok && c.(*watcherPath).send { //配置已删除
				w.cacheAddress.Remove(path)
				if c.(*watcherPath).conf != nil {
					updater := &server.Updater{Conf: c.(*watcherPath).conf, Op: conf.DEL}
					w.notifyConfChan <- updater
				}
			}
			break START
		case <-time.After(w.timeSpan):
			if w.done {
				break START
			}

			b := w.checker.Exists(path)
			if !b {
				ch <- struct{}{}
				continue
			}

			modify, err := w.checker.LastModeTime(path)
			if err != nil {
				continue
			}
			updater := &server.Updater{}
			cc, ok := w.cacheAddress.Get(path)
			v := cc.(*watcherPath)
			if !ok {
				break START
			} else {
				if v.modTime == w.defTime {
					updater.Op = conf.ADD
				} else if v.modTime != modify {
					updater.Op = conf.CHANGE //检查配置项变化
				} else {
					continue
				}
			}
			updater.Conf, err = w.getConf(path)
			if err != nil {
				continue
			}
			v.modTime = modify
			v.conf = updater.Conf
			v.send = true
			w.notifyConfChan <- updater
		}
	}
	return
}

//getConf 获取配置
func (w *jsonConfWatcher) getConf(path string) (cf conf.Conf, err error) {
	buf, err := w.checker.ReadAll(path)
	if err != nil {
		return
	}
	if len(buf) < 3 {
		return nil, errors.New("配置文件为空")
	}
	c := make(map[string]interface{})
	err = json.Unmarshal(buf, &c)
	if err != nil {
		return
	}
	return conf.NewJSONConf(c), nil
}

//Close 关闭所有监控项
func (w *jsonConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	return nil
}

//Notify 节点变化后通知
func (w *jsonConfWatcher) Notify() chan *server.Updater {
	return w.notifyConfChan
}
