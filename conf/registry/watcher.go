package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

//Registry 注册中心接口
type Registry interface {
	Exists(path string) bool
	WatchChildren(path string) (data chan []string, err error)
	WatchValue(path string) (data chan string, err error)
	GetChildren(path string) (data []string, err error)
	GetValue(path string) (data string, err error)
}

type registryConfWatcher struct {
	watchConfChan  chan string
	delConfChan    chan string
	watchRootChan  chan string
	cacheAddress   cmap.ConcurrentMap
	cacheDir       cmap.ConcurrentMap
	notifyConfChan chan *conf.Updater
	defTime        time.Time
	isInitialized  bool
	done           bool
	registry       Registry
	timeSpan       time.Duration
	mu             sync.Mutex
	domain         string
	tag            string
}

type watcherPath struct {
	close chan struct{}
	conf  conf.Conf
	send  bool
}

//NewRegistryConfWatcher 创建zookeeper配置文件监控器
func NewRegistryConfWatcher(domain string, tag string, registry Registry) (w *registryConfWatcher) {
	w = &registryConfWatcher{
		notifyConfChan: make(chan *conf.Updater),
		watchConfChan:  make(chan string, 2),
		delConfChan:    make(chan string, 2),
		watchRootChan:  make(chan string, 10),
		cacheAddress:   cmap.New(),
		cacheDir:       cmap.New(),
		registry:       registry,
		timeSpan:       time.Second,
		domain:         domain,
		tag:            tag,
	}
	w.defTime, _ = time.Parse("2000-01-01", "2006-01-02")
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
	for _, v := range conf.WatchServers {
		path := fmt.Sprintf("%s/%s", w.domain, v)
		w.cacheDir.Set(path, false)
		w.watchRootChan <- path
	}
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
			go w.watchPath(p)
		case p, ok := <-w.watchConfChan:
			if w.done || !ok {
				break START
			}
			go w.watchConf(p)
		case del := <-w.delConfChan:
			w.cacheAddress.IterCb(func(key string, value interface{}) bool {

				if strings.HasPrefix(key, del) {
					fmt.Println(":::", key, del)
					value.(*watcherPath).close <- struct{}{}
				}
				return true
			})
		}
	}
}

//watchPath 监控配置路径变化, 文件夹不存在时判断是否已删除，已删除则关闭所有子目录监控，否则获取所有子目录并添加到监控目录
func (w *registryConfWatcher) watchPath(path string) (err error) {
	fmt.Println("check exists:", path)
	//持续监控节点是否存在
	isExists := w.registry.Exists(path)
WATCH_EXISTS:
	for !isExists {
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				break WATCH_EXISTS
			}
			if isExists = w.registry.Exists(path); !isExists {
				w.removePath(path)
			}
		}
	}
	if !isExists {
		return fmt.Errorf("路径不存在:%s", path)
	}
	fmt.Println("get chilren:", path)
	//获取所有子节点
	children, err := w.registry.GetChildren(path)
	if err != nil {
		return
	}
	w.cacheDir.Set(path, true)
	w.notifyPathChange(path, children)

	//监控子节点变化
	ch, err := w.registry.WatchChildren(path)
	if err != nil {
		return
	}

WATCH_CHILDREN:
	for {
		select {
		case children, ok := <-ch:
			if !ok || w.done {
				break WATCH_CHILDREN
			}
			w.notifyPathChange(path, children)
			//继续监控子节点变化
			ch, err = w.registry.WatchChildren(path)
			if err != nil {
				return
			}
		}
	}
	return
}

//watchConf 监控配置项变化，当配置文件删除后关闭监控并发送删除通知，发生变化则只发送变更通知
func (w *registryConfWatcher) watchConf(path string) (err error) {
	fmt.Println("watch conf:", path)
	//持续监控节点是否存在
	c, _ := w.cacheAddress.Get(path)
	closeCh := c.(*watcherPath).close
	isExists := w.registry.Exists(path)
START1:
	for !isExists {
		select {
		case <-closeCh:
			if w.done {
				break START1
			}
			w.removeConf(path)
			break START1
		case <-time.After(w.timeSpan):
			if w.done {
				break START1
			}
			isExists = w.registry.Exists(path)
		}
	}
	if !isExists {
		return errors.New("路径不存在")
	}
	fmt.Println("get value:", path)
	//获取节点值
	data, err := w.registry.GetValue(path)
	if err != nil {
		return
	}
	c.(*watcherPath).send = true
	if err = w.notifyConfChange(path, data, conf.ADD); err != nil {
		return
	}
	fmt.Println("watch value:", path)
	//持续监控节点值变化
	dataChan, err := w.registry.WatchValue(path)
	if err != nil {
		return
	}

START:
	for {
		select {
		case <-closeCh:
			fmt.Println("close ch:", path)
			if w.done {
				break START
			}
			w.removeConf(path)
			break START
		case content := <-dataChan:
			fmt.Println("watch value 1:", path)
			if w.done {
				break START
			}

			if err = w.notifyConfChange(path, content, conf.CHANGE); err != nil {
				continue
			}

			//继续监控值变化
			fmt.Println("watch value 2:", path)
			dataChan, err = w.registry.WatchValue(path)
			if err != nil {
				return
			}

		}
	}
	return
}
func (w *registryConfWatcher) removePath(path string) {
	if v, ok := w.cacheDir.Get(path); ok && v.(bool) { //目录已删除
		fmt.Println("remove path:", path)
		w.delConfChan <- path
		w.cacheDir.Set(path, false)
	}
}
func (w *registryConfWatcher) removeConf(path string) {
	if c, ok := w.cacheAddress.Get(path); ok && c.(*watcherPath).send { //配置已删除
		w.cacheAddress.Remove(path)
		if c.(*watcherPath).conf != nil {
			updater := &conf.Updater{Conf: c.(*watcherPath).conf, Op: conf.DEL}
			w.notifyConfChan <- updater
		}
	}
}
func (w *registryConfWatcher) notifyConfChange(path string, content string, op int) (err error) {
	updater := &conf.Updater{Op: op}
	cc, ok := w.cacheAddress.Get(path)
	if !ok {
		return
	}
	v := cc.(*watcherPath)
	updater.Conf, err = w.getConf(content)
	if err != nil {
		return
	}
	v.conf = updater.Conf
	w.notifyConfChan <- updater
	return
}
func (w *registryConfWatcher) notifyPathChange(path string, children []string) {
	for _, v := range children { //检查当前配置地址未缓存
		name := fmt.Sprintf("%s/%s/conf/%s", path, v, w.tag)
		if _, ok := w.cacheAddress.Get(name); !ok {
			w.cacheAddress.Set(name, &watcherPath{close: make(chan struct{}, 1)})
			w.watchConfChan <- name
		}
	}
	w.cacheAddress.IterCb(func(key string, value interface{}) bool {
		exists := false
		for _, v := range children {
			fmt.Println(">>:", key, fmt.Sprintf("%s/%s/conf/%s", path, v, w.tag), key == fmt.Sprintf("%s/%s/conf/%s", path, v, w.tag))
			if key == fmt.Sprintf("%s/%s/conf/%s", path, v, w.tag) {
				exists = true
				break
			}
		}
		if !exists {
			fmt.Println(">> close:", path)
			value.(*watcherPath).close <- struct{}{}
		}
		return false
	})
}

//getConf 获取配置
func (w *registryConfWatcher) getConf(content string) (cf conf.Conf, err error) {
	c := make(map[string]interface{})
	err = json.Unmarshal([]byte(content), &c)
	if err != nil {
		return
	}
	return conf.NewJSONConf(c), nil
}

//Close 关闭所有监控项
func (w *registryConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	return nil
}

//Notify 节点变化后通知
func (w *registryConfWatcher) Notify() (chan *conf.Updater, error) {
	return w.notifyConfChan, nil
}
