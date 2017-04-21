package cluster

import (
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"sync"

	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
)

type watchConf struct {
	path           string
	registry       registry.Registry
	timeSpan       time.Duration
	notifyConfChan chan *conf.Updater
	conf           conf.Conf
	serverName     string
	done           bool
	notifyCount    int32
	mu             sync.Mutex
	category       string
	domain         string
	args           map[string]string
}

func NewWatchConf(domain string, serverName string, category string, path string, registry registry.Registry, updater chan *conf.Updater, timeSpan time.Duration) *watchConf {
	return &watchConf{path: path,
		domain:         domain,
		registry:       registry,
		serverName:     serverName,
		category:       category,
		notifyConfChan: updater,
		timeSpan:       timeSpan,
		args:           make(map[string]string),
	}
}

//watchConf 监控配置项变化，当发生错误时持续监控节点变化，只有明确节点不存在时才会通知关闭
func (w *watchConf) watch() (err error) {
	//持续监控节点是否存在
LOOP:
	isExists, _ := w.registry.Exists(w.path)
	for !isExists {
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				return errors.New("watcher is closing")
			}
			isExists, err = w.registry.Exists(w.path)
			if !isExists && err == nil {
				w.NotifyConfDel()
			}
		}
	}

	//获取节点值
	data, version, err := w.registry.GetValue(w.path)
	if err != nil {
		goto LOOP
	}
	if err = w.notifyConfChange(data, version); err != nil {
		time.Sleep(w.timeSpan)
		goto LOOP
	}
	dataChan, err := w.registry.WatchValue(w.path)
	if err != nil {
		goto LOOP
	}

	for {
		select {
		case <-time.After(w.timeSpan):
			if w.done {
				return errors.New("watcher is closing")
			}
		case content, ok := <-dataChan:
			if w.done || !ok {
				return errors.New("watcher is closing")
			}
			if err = content.GetError(); err != nil {
				goto LOOP
			}
			w.notifyConfChange(content.GetValue())

			//继续监控值变化
			dataChan, err = w.registry.WatchValue(w.path)
			if err != nil {
				goto LOOP
			}
		}
	}
}

func (w *watchConf) notifyConfChange(content []byte, version int32) (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	op := registry.ADD
	if atomic.LoadInt32(&w.notifyCount) > 0 {
		op = registry.CHANGE
	}
	updater := &conf.Updater{Op: op}
	updater.Conf, err = w.getConf(content, version)
	if err != nil {
		return
	}
	w.conf = updater.Conf
	w.notifyConfChan <- updater
	atomic.AddInt32(&w.notifyCount, 1)
	return
}

//getConf 获取配置
func (w *watchConf) getConf(content []byte, version int32) (cf conf.Conf, err error) {
	c := make(map[string]interface{})
	err = json.Unmarshal(content, &c)
	if err != nil {
		return
	}
	for k, v := range w.args {
		c[k] = v
	}
	jconf := conf.NewJSONConfWithHandle(c, version, w.getValue)
	jconf.Set("name", w.serverName)
	jconf.Set("domain", w.domain)
	jconf.Set("path", w.path)
	jconf.Set("root_path", fmt.Sprintf("%s/servers/%s/%s/conf", w.domain, w.serverName, w.category))
	jconf.Set("category_path", fmt.Sprintf("%s/servers/%s/%s", w.domain, w.serverName, w.category))
	jconf.Set("server_path", fmt.Sprintf("%s/servers/%s", w.domain, w.serverName))
	jconf.Content = string(content)
	return jconf, nil
}
func (w *watchConf) getValue(path string) (r conf.Conf, err error) {
	buf, version, err := w.registry.GetValue(path)
	if err != nil {
		return
	}
	return w.getConf(buf, version)
}
func (w *watchConf) NotifyConfDel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil { //配置已删除
		updater := &conf.Updater{Conf: w.conf, Op: registry.DEL}
		w.notifyConfChan <- updater
		w.conf = nil
		atomic.AddInt32(&w.notifyCount, -1)
	}
}
func (w *watchConf) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
}
