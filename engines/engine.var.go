package engines

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"

	"github.com/qxnw/hydra/registry"
)

type varValue struct {
	version  int32
	path     string
	registry registry.Registry
	value    []byte
	timeSpan time.Duration
	*varParamWatcher
}

type varParamWatcher struct {
	domain   string
	path     string
	registry registry.Registry
	wathcer  cmap.ConcurrentMap
	once     sync.Once

	done      bool
	closeChan chan struct{}
	lock      sync.Mutex
}

func newVarParamWatcher(domain string, r registry.Registry) *varParamWatcher {
	return &varParamWatcher{domain: domain, registry: r, wathcer: cmap.New(6), closeChan: make(chan struct{})}
}
func newVarValue(path string, r registry.Registry, v *varParamWatcher) *varValue {
	return &varValue{path: path, registry: r, timeSpan: time.Second, varParamWatcher: v}
}
func (v *varValue) getValue() ([]byte, int32, error) {
	if v.version == 0 {
		return v.registry.GetValue(v.path)
	}
	return v.value, v.version, nil
}
func (v *varValue) notify(value []byte, version int32) {
	v.version = version
	v.value = value
}
func (v *varValue) watch() error {
LOOP:
	isExists, _ := v.registry.Exists(v.path)
	for !isExists {
		select {
		case <-time.After(v.timeSpan):
			if v.done {
				return errors.New("watcher is closing")
			}
			isExists, err := v.registry.Exists(v.path)
			if !isExists && err == nil {
				v.notify(nil, 0)
			}
		}
	}

	//获取节点值
	data, version, err := v.registry.GetValue(v.path)
	if err != nil {
		goto LOOP
	}
	v.notify(data, version)
	dataChan, err := v.registry.WatchValue(v.path)
	if err != nil {
		goto LOOP
	}
LOOP2:
	for {
		select {
		case <-v.closeChan:
			return errors.New("watcher is closing")
		case content, ok := <-dataChan:
			if v.done || !ok {
				return errors.New("watcher is closing")
			}
			if err = content.GetError(); err != nil {
				goto LOOP
			}
			v.notify(content.GetValue())
			//继续监控值变化
			dataChan, err = v.registry.WatchValue(v.path)
			if err != nil {
				goto LOOP2
			}
		}
	}
}
func (r *varParamWatcher) GetVarValue(tp string, name string) ([]byte, int32, error) {
	path := fmt.Sprintf("/%s/var/%s/%s", r.domain, tp, name)
	_, cached, _ := r.wathcer.SetIfAbsentCb(path, func(input ...interface{}) (c interface{}, err error) {
		path := input[0].(string)
		varValue := newVarValue(path, r.registry, r)
		go varValue.watch()
		return varValue, nil

	}, path)
	c := cached.(*varValue)
	return c.getValue()
}
func (r *varParamWatcher) Close() error {
	r.done = true
	r.once.Do(func() {
		close(r.closeChan)
	})
	return nil
}

//GetVarParam 获取配置参数
func (r *varParamWatcher) GetVarParam(tp string, name string) (string, error) {
	buff, _, err := r.GetVarValue(tp, name)
	if err != nil {
		return "", err
	}
	return string(buff), nil
}
