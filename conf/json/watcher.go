package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/qxnw/hydra/conf"
)

type jsonConfWatcher struct {
	watchConfChan  chan string
	deleteConfChan chan string
	watchRootChan  chan string
	cacheAddress   map[string]time.Time
	cacheConf      map[string]conf.Conf
	notifyConfChan chan conf.Updater
	isInitialized  bool
	done           bool
	mu             sync.Mutex
	checker        checker
	timeSpan       time.Duration
}

//NewJSONConfWatcher 创建zookeeper配置文件监控器
func NewJSONConfWatcher() *jsonConfWatcher {
	return &jsonConfWatcher{
		notifyConfChan: make(chan conf.Updater, 1),
		watchConfChan:  make(chan string, 1),
		deleteConfChan: make(chan string, 1),
		watchRootChan:  make(chan string, 10),
		cacheAddress:   make(map[string]time.Time),
		cacheConf:      make(map[string]conf.Conf),
		checker:        fileChecker{},
		timeSpan:       time.Second,
	}
}
func (w *jsonConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	for _, v := range conf.WatchServers {
		w.watchRootChan <- fmt.Sprintf("../%s", v)
	}
	go w.watch()
	return nil
}
func (w *jsonConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	return nil
}

//Notify 节点变化后通知
func (w *jsonConfWatcher) Notify() (chan conf.Updater, error) {
	return w.notifyConfChan, nil
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

//watchPath 监控配置路径变化
func (w *jsonConfWatcher) watchPath(path string) (err error) {
START:
	for {
		select {
		case <-time.After(w.timeSpan):
			w.mu.Lock()
			if w.done {
				w.mu.Unlock()
				break START
			}
			b := w.checker.Exists(path)
			if !b {
				w.mu.Unlock()
				continue
			}
			children, err := w.checker.ReadDir(path)
			if err != nil {
				w.mu.Unlock()
				continue
			}
			for _, v := range children { //检查当前配置地址未缓存
				name := fmt.Sprintf("%s/%s/conf/conf.json", path, v)
				if _, ok := w.cacheAddress[name]; !ok {
					w.watchConfChan <- name
				}
			}
			for k := range w.cacheAddress { //当前地址已删除
				exists := false
				for _, v := range children {
					name := fmt.Sprintf("%s/conf/conf.json", v)
					if k == name {
						exists = true
						break
					}
				}
				if !exists {
					fmt.Println("deleteConfChan:1")
					w.deleteConfChan <- k
				}
			}
			w.mu.Unlock()
		case del := <-w.deleteConfChan:
			w.mu.Lock()
			if w.done {
				w.mu.Unlock()
				break START
			}
			fmt.Printf("delete..%s,%+v\n", del, w.cacheConf)
			if c, ok := w.cacheConf[del]; ok { //配置已删除
				fmt.Println("delete now..", del)
				w.notifyConfChan <- conf.Updater{Conf: c, Op: conf.DEL}
				delete(w.cacheConf, del)
			}
			w.mu.Unlock()
		}
	}
	return
}

//watchConf 监控配置项变化，变化后发送通知
func (w *jsonConfWatcher) watchConf(path string) (err error) {
START:
	for {
		select {
		case <-time.After(w.timeSpan):
			w.mu.Lock()
			if w.done {
				w.mu.Unlock()
				break START
			}
			modify, err := w.checker.LastModeTime(path)
			if err != nil {
				w.mu.Unlock()
				continue
			}

			updater := conf.Updater{}
			if v, ok := w.cacheAddress[path]; !ok {
				updater.Op = conf.ADD //检查配置项变化
			} else if modify != v {
				updater.Op = conf.CHANGE //检查配置项变化
			} else {
				w.mu.Unlock()
				continue
			}
			updater.Conf, err = w.getConf(path)
			if err != nil {
				w.mu.Unlock()
				fmt.Println("deleteConfChan:2", path)
				w.deleteConfChan <- path //??
				continue
			}
			if updater.Op == conf.ADD || updater.Op == conf.CHANGE {
				w.cacheConf[path] = updater.Conf
			}

			w.cacheAddress[path] = modify
			w.notifyConfChan <- updater
			w.mu.Unlock()
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
