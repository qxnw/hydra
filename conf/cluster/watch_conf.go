package cluster

import (
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"sync"

	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

type watchConf struct {
	confPath       string
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
	closeChan      chan struct{}
	tagName        string
	*logger.Logger
}

//newWatchConf 监控配置文件变化
func newWatchConf(domain string, serverName string, category string, tagName string, confPath string, registry registry.Registry,
	updater chan *conf.Updater, timeSpan time.Duration, log *logger.Logger) *watchConf {
	return &watchConf{confPath: confPath,
		domain:         strings.Trim(domain, "/"),
		registry:       registry,
		serverName:     serverName,
		category:       category,
		tagName:        tagName,
		notifyConfChan: updater,
		timeSpan:       timeSpan,
		Logger:         log,
		args:           make(map[string]string),
		closeChan:      make(chan struct{}),
	}
}


func (w *watchConf) notifyDeleted() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil { //配置已删除
		updater := &conf.Updater{Conf: w.conf, Op: registry.DEL}
		w.notifyConfChan <- updater
		w.conf = nil
		atomic.AddInt32(&w.notifyCount, -1)
	}
}
func (w *watchConf) notifyChanged(content []byte, version int32) (err error) {
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
	jconf := conf.NewJSONConfWithHandle(c, version, w.registry)
	jconf.Set("name", w.serverName)
	jconf.Set("domain", w.domain)
	jconf.Set("path", w.confPath)
	jconf.Set("type", w.category)
	jconf.Set("tag", w.tagName)

	//	jconf.Set("root_path", fmt.Sprintf("/%s/conf/%s/%s", w.domain, w.serverName, w.category))
	//jconf.Set("category_path", fmt.Sprintf("/%s/conf/%s/%s", w.domain, w.serverName, w.category))
	//jconf.Set("server_path", fmt.Sprintf("/%s/conf/%s", w.domain, w.serverName))

	jconf.Set("root_path", fmt.Sprintf("/%s/%s/%s/conf", w.domain, w.serverName, w.category))
	jconf.Set("category_path", fmt.Sprintf("/%s/%s/%s/conf", w.domain, w.serverName, w.category))
	jconf.Set("server_path", fmt.Sprintf("/%s/%s", w.domain, w.serverName))

	jconf.Content = string(content)
	return jconf, nil
}

/*
func (w *watchConf) getValue2(path string) (buf []byte, err error) {
	buf, _, err = w.registry.GetValue(path)
	if err != nil {
		return
	}
	return
}
func (w *watchConf) getValue(path string) (r conf.Conf, err error) {
	buf, version, err := w.registry.GetValue(path)
	if err != nil {
		return
	}
	return w.getConf(buf, version)
}*/

func (w *watchConf) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	close(w.closeChan)
}
