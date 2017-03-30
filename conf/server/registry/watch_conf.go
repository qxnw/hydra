package registry

import (
	"encoding/json"
	"errors"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/conf/server"
)

type watchConf struct {
	path           string
	registry       Registry
	timeSpan       time.Duration
	notifyConfChan chan *server.Updater
	conf           conf.Conf
	done           bool
	notifyCount    int
	mu             sync.Mutex
}

func NewWatchConf(path string, registry Registry, updater chan *server.Updater, timeSpan time.Duration) *watchConf {
	return &watchConf{path: path, registry: registry, notifyConfChan: updater, timeSpan: timeSpan}
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
	data, err := w.registry.GetValue(w.path)
	if err != nil {
		goto LOOP
	}
	if err = w.notifyConfChange(data); err != nil {
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

func (w *watchConf) notifyConfChange(content []byte) (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	op := conf.ADD
	if w.notifyCount > 0 {
		op = conf.CHANGE
	}
	updater := &server.Updater{Op: op}
	updater.Conf, err = w.getConf(content)
	if err != nil {
		return
	}
	w.conf = updater.Conf
	w.notifyConfChan <- updater
	w.notifyCount++
	return
}

//getConf 获取配置
func (w *watchConf) getConf(content []byte) (cf conf.Conf, err error) {
	c := make(map[string]interface{})
	err = json.Unmarshal(content, &c)
	if err != nil {
		return
	}
	return conf.NewJSONConf(c), nil
}

func (w *watchConf) NotifyConfDel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil { //配置已删除
		updater := &server.Updater{Conf: w.conf, Op: conf.DEL}
		w.notifyConfChan <- updater
	}
}
func (w *watchConf) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
}
