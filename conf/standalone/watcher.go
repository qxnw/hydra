package standalone

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"strings"

	"sync"

	"os"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
)

type jsonConfWatcher struct {
	watchConfChan  chan string
	deleteConfChan chan string
	watchRootChan  chan string
	cacheAddress   cmap.ConcurrentMap
	cacheDir       cmap.ConcurrentMap
	notifyConfChan chan *conf.Updater
	defTime        time.Time
	isInitialized  bool
	done           bool
	checker        registry.Checker
	timeSpan       time.Duration
	domain         string
	serverTag      string
	mu             sync.Mutex
	closeChan      chan struct{}
	*logger.Logger
}

type watcherPath struct {
	close        chan struct{}
	serverName   string
	modTime      time.Time
	conf         conf.Conf
	categoryPath string
	serverPath   string
	confRoot     string
	typeName     string
	send         bool
}

//newSAWatcher 创建基于本地文件的配置监控器
func newSAWatcher(domain string, serverTag string, log *logger.Logger) (w *jsonConfWatcher, err error) {
	w = &jsonConfWatcher{
		notifyConfChan: make(chan *conf.Updater),
		watchConfChan:  make(chan string, 2),
		deleteConfChan: make(chan string, 2),
		watchRootChan:  make(chan string, 10),
		closeChan:      make(chan struct{}),
		cacheAddress:   cmap.New(2),
		cacheDir:       cmap.New(2),
		domain:         domain,
		serverTag:      serverTag,
		timeSpan:       time.Second,
		Logger:         log,
	}
	w.checker, err = registry.NewChecker()
	if err != nil {
		return
	}
	if serverTag == "" {
		w.serverTag = "conf"
	}
	w.defTime, _ = time.Parse("2000-01-01", "2006-01-02")
	return
}
func (w *jsonConfWatcher) getPathPrefix() string {
	prefix := ""
	if runtime.GOOS == "windows" {
		p := os.Args[0]
		prefix = string(p[0]) + ":"
	}
	return prefix
}

//Start 启用配置文件监控
func (w *jsonConfWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isInitialized {
		return nil
	}
	w.isInitialized = true
	path := fmt.Sprintf("%s%s/servers", w.getPathPrefix(), w.domain)
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
		case <-w.closeChan:
			break START
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
		case <-w.closeChan:
			break START
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
				for _, sv := range conf.WatchServers { //hydra/servers/merchant.api/api/conf.json
					name := fmt.Sprintf("%s/%s/%s/conf/conf", path, v, sv)
					if _, ok := w.cacheAddress.Get(name); !ok {
						w.cacheAddress.Set(name, &watcherPath{modTime: w.defTime,
							serverName:   v,
							typeName:     sv,
							confRoot:     fmt.Sprintf("%s/%s/%s/conf", path, v, sv),
							categoryPath: fmt.Sprintf("%s/%s/%s", path, v, sv),
							serverPath:   fmt.Sprintf("%s/%s", path, v),
							close:        make(chan struct{}, 1)})
						w.watchConfChan <- name
					}
				}
			}
			w.cacheAddress.IterCb(func(key string, value interface{}) bool {
				exists := false
				for _, v := range children {
					for _, sv := range conf.WatchServers {
						if key == fmt.Sprintf("%s/%s/%s/conf/conf", path, v, sv) {
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
		case <-w.closeChan:
			break START
		case <-ch:
			if w.done {
				break START
			}
			if c, ok := w.cacheAddress.Get(path); ok && c.(*watcherPath).send { //配置已删除
				w.cacheAddress.Remove(path)
				if c.(*watcherPath).conf != nil {
					updater := &conf.Updater{Conf: c.(*watcherPath).conf, Op: registry.DEL}
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
				//fmt.Println("not exist:", path)
				//ch <- struct{}{}
				continue
			}

			modify, err := w.checker.LastModeTime(path)
			if err != nil {
				continue
			}
			updater := &conf.Updater{}
			cc, ok := w.cacheAddress.Get(path)
			v := cc.(*watcherPath)
			if !ok {
				break START
			} else {
				if v.modTime == w.defTime {
					updater.Op = registry.ADD
				} else if v.modTime != modify {
					updater.Op = registry.CHANGE //检查配置项变化
				} else {
					continue
				}
			}
			cfx, err := w.getConf(path)
			if err != nil {
				w.Warnf("节点配置错误：%s(err:%v)", path, err)
				time.Sleep(time.Second * 5)
				continue
			}
			updater.Conf = cfx.(conf.Conf)
			v.modTime = modify
			v.conf = updater.Conf
			v.send = true
			w.notifyConfChan <- updater
		}
	}
	return
}
func (w *jsonConfWatcher) getValue(path string) (r []byte, err error) {
	buf, err := w.checker.ReadAll(path)
	if err != nil {
		return
	}
	if len(buf) < 3 {
		return nil, errors.New("配置文件为空")
	}
	return buf, nil
}

//getConf 获取配置
func (w *jsonConfWatcher) getConf(path string) (cf conf.Conf, err error) {
	f, err := os.Stat(path)
	if err != nil {
		return
	}
	buf, err := w.getValue(path)
	if err != nil {
		return
	}
	c := make(map[string]interface{})
	err = json.Unmarshal(buf, &c)
	if err != nil {
		return
	}
	c["domain"] = w.domain

	jcf := conf.NewJSONConfWithHandle(c, int32(f.ModTime().Unix()), nil) //8.24 colin 不再支持单机模式

	if cc, ok := w.cacheAddress.Get(path); ok {
		v := cc.(*watcherPath)
		jcf.Set("root_path", v.confRoot)
		jcf.Set("path", v.confRoot)
		jcf.Set("category_path", v.categoryPath)
		jcf.Set("server_path", v.serverPath)
		jcf.Set("name", v.serverName)
		jcf.Set("type", v.typeName)
	}
	jcf.Content = string(buf)
	return jcf, nil
}

//Close 关闭所有监控项
func (w *jsonConfWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	close(w.closeChan)
	return nil
}

//Notify 节点变化后通知
func (w *jsonConfWatcher) Notify() chan *conf.Updater {
	return w.notifyConfChan
}
