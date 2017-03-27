package zookeeper

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"strings"

	"errors"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/zk"
)

type zookeeperConfWatcher struct {
	client         *zk.ZookeeperClient
	notifyConfChan chan conf.Updater
	watchConfChan  chan string
	deleteConfChan chan string
	watchRootChan  chan string
	cacheAddress   map[string]string
	cacheConf      map[string]conf.Conf
	domain         string
	isInitialized  bool
	done           bool
	tag            string
	mu             sync.Mutex
}

//NewZookeeperConfWatcher 创建zookeeper配置文件监控器
func NewZookeeperConfWatcher(domain string, tag string, client *zk.ZookeeperClient) conf.ConfWatcher {
	return &zookeeperConfWatcher{
		client:         client,
		notifyConfChan: make(chan conf.Updater, 1),
		watchConfChan:  make(chan string, 1),
		deleteConfChan: make(chan string, 1),
		watchRootChan:  make(chan string, 10),
		cacheAddress:   make(map[string]string),
		cacheConf:      make(map[string]conf.Conf),
		domain:         strings.Trim(domain, "/"),
		tag:            tag,
	}
}

//Start 启用Watch, 当配置发生变化后自动发送通知
func (w *zookeeperConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	for _, v := range conf.WatchServers {
		rootPath := fmt.Sprintf("/%s/%s", w.domain, v)
		w.watchRootChan <- rootPath
	}
	go w.watch()
	return nil
}
func (w *zookeeperConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	w.client.Disconnect()
	return nil
}

//watch 监控配置路径变化和配置数据变化
func (w *zookeeperConfWatcher) watch() {
START:
	for {
		select {
		case <-time.After(time.Second):
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
func (w *zookeeperConfWatcher) watchPath(path string) (err error) {
	_, err = w.client.WaitExists(path)
	if err != nil {
		return
	}
	data := make(chan []string, 1)
	err = w.client.BindWatchChildren(path, data)
	if err != nil {
		return
	}
START:
	for {
		select {
		case <-time.After(time.Second):
			if w.done {
				break START
			}
		case children, ok := <-data:
			w.mu.Lock()
			if w.done || !ok {
				w.mu.Unlock()
				break START
			}
			for _, v := range children { //检查当前配置地址未缓存
				name := fmt.Sprintf("/%s/conf/%s", v, w.tag)
				if _, ok := w.cacheAddress[name]; !ok {
					w.cacheAddress[name] = name
					w.watchConfChan <- name
				}
			}
			for k := range w.cacheAddress { //当前地址已删除
				exists := false
				for _, v := range children {
					name := fmt.Sprintf("/%s/conf/%s", v, w.tag)
					if k == name {
						exists = true
						break
					}
				}
				if !exists {
					w.deleteConfChan <- k
				}
			}
			w.mu.Unlock()
		case del := <-w.deleteConfChan:
			if w.done {
				break START
			}
			w.mu.Lock()
			if c, ok := w.cacheConf[del]; ok { //配置已删除
				w.notifyConfChan <- conf.Updater{Conf: c, Op: conf.DEL}
				delete(w.cacheConf, del)
			}
			w.mu.Unlock()
		}
	}
	return
}

//watchConf 监控配置项变化，变化后发送通知
func (w *zookeeperConfWatcher) watchConf(path string) (err error) {
	data := make(chan string, 1)
	err = w.client.BindWatchValue(path, data)
	if err != nil {
		return
	}

START:
	for {
		select {
		case <-time.After(time.Second):
			if w.done {
				break START
			}
		case value := <-data:
			if w.done {
				break START
			}
			updater := conf.Updater{}
			updater.Conf, err = w.getConf(value)
			if err != nil {
				continue
			}
			w.mu.Lock()
			if _, ok := w.cacheConf[path]; ok {
				updater.Op = conf.CHANGE //检查配置项变化
			} else {
				updater.Op = conf.ADD //检查配置项变化
				w.cacheConf[path] = updater.Conf
			}
			w.mu.Unlock()
			w.notifyConfChan <- updater
		}
	}
	return
}

//Notify 节点变化后通知
func (w *zookeeperConfWatcher) Notify() (chan conf.Updater, error) {
	return w.notifyConfChan, nil
}

//getConf 获取配置
func (w *zookeeperConfWatcher) getConf(data string) (cf conf.Conf, err error) {
	if data == "" {
		return nil, errors.New("配置文件为空")
	}
	c := make(map[string]interface{})
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return
	}
	return conf.NewJSONConf(c), nil
}
