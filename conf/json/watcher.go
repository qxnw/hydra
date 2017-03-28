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
	watchingMap    map[string]chan struct{}
	deleteConfChan chan string
	watchRootChan  chan string
	cacheAddress   map[string]time.Time
	cacheConf      map[string]conf.Conf
	cacheDir       map[string]bool
	notifyConfChan chan *conf.Updater
	defTime        time.Time
	isInitialized  bool
	done           bool
	mu             sync.Mutex
	checker        checker
	timeSpan       time.Duration
}

//NewJSONConfWatcher 创建zookeeper配置文件监控器
func NewJSONConfWatcher() (w *jsonConfWatcher) {
	w = &jsonConfWatcher{
		watchingMap:    make(map[string]chan struct{}),
		notifyConfChan: make(chan *conf.Updater),
		watchConfChan:  make(chan string, 2),
		deleteConfChan: make(chan string, 2),
		watchRootChan:  make(chan string, 10),
		cacheAddress:   make(map[string]time.Time),
		cacheDir:       make(map[string]bool),
		cacheConf:      make(map[string]conf.Conf),
		checker:        fileChecker{},
		timeSpan:       time.Second,
	}
	w.defTime, _ = time.Parse("2000-01-01", "2006-01-02")
	return

}
func (w *jsonConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	for _, v := range conf.WatchServers {
		path := fmt.Sprintf("../%s", v)
		w.cacheDir[path] = false
		w.watchRootChan <- path
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
func (w *jsonConfWatcher) Notify() (chan *conf.Updater, error) {
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
		case del := <-w.deleteConfChan:
			w.mu.Lock()
			fmt.Println("delete:chan")
			fmt.Printf("dddd:%v\n", len(w.notifyConfChan))
			if ch, ok := w.watchingMap[del]; ok {
				ch <- struct{}{}
				fmt.Println("add to delete ch")
				delete(w.watchingMap, del)
			}
			w.mu.Unlock()
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
				if w.cacheDir[path] {

				}
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
					w.cacheAddress[name] = w.defTime
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

		}
	}
	return
}

//watchConf 监控配置项变化，变化后发送通知
func (w *jsonConfWatcher) watchConf(path string) (err error) {
	w.mu.Lock()
	if _, ok := w.watchingMap[path]; ok {
		w.mu.Unlock()
		return
	}

	ch := make(chan struct{}, 10)
	w.watchingMap[path] = ch
	w.mu.Unlock()
START:
	for {
		select {
		case <-ch:
			w.mu.Lock()
			if w.done {
				w.mu.Unlock()
				break START
			}
			if c, ok := w.cacheConf[path]; ok { //配置已删除
				fmt.Println("delelte:", path)
				updater := &conf.Updater{Conf: c, Op: conf.DEL}
				w.notifyConfChan <- updater
				fmt.Printf("dddd:%+v\n", len(w.notifyConfChan))
			}
			w.mu.Unlock()
			break START
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

			updater := &conf.Updater{}
			if v, ok := w.cacheAddress[path]; ok && v == w.defTime {
				w.cacheAddress[path] = modify
				updater.Op = conf.ADD //检查配置项变化
			} else if modify.Sub(v) != 0 {
				fmt.Printf("%+v----%+v", modify, v)
				w.cacheAddress[path] = modify
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
