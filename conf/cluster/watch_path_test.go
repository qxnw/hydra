package cluster

import (
	"testing"
	"time"

	"errors"

	"github.com/qxnw/hydra/registry/conf"
	rx "github.com/qxnw/lib4go/registry"
)

type pathRegistry struct {
	watchChan chan rx.ChildrenWatcher
	watchErr  error
	exists    map[string]bool
	children  []string
}

func (r *pathRegistry) Exists(path string) (bool, error) {
	if v, ok := r.exists[path]; ok {
		return v, nil
	}
	return false, nil
}
func (r *pathRegistry) WatchValue(path string) (data chan rx.ValueWatcher, err error) {
	return nil, nil
}

func (r *pathRegistry) GetValue(path string) (data []byte, err error) {
	return nil, nil
}

func (r *pathRegistry) WatchChildren(path string) (data chan rx.ChildrenWatcher, err error) {
	data = r.watchChan
	err = r.watchErr
	return
}
func (r *pathRegistry) GetChildren(path string) (data []string, err error) {
	data = r.children
	return
}
func (r *pathRegistry) CreatePersistentNode(path string, data string) (err error) {
	return nil
}
func (r *pathRegistry) CreateTempNode(path string, data string) (err error) {
	return nil
}
func (r *pathRegistry) CreateSeqNode(path string, data string) (rpath string, err error) {
	return "", nil
}
func TestPathWatcher1(t *testing.T) {
	r := &pathRegistry{
		watchChan: make(chan rx.ChildrenWatcher, 1),
	}
	updater := make(chan *conf.Updater, 4)
	watcher := NewWatchPath("/hydra", "192.168.0.1:001", "/hydra/servers", r, updater, time.Millisecond*10)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/servers": true,
	}
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}
	r.watchChan <- &valuesEntity{values: []string{"merchant"}}
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.exists, true)
	expect(t, watcher.cacheAddress.Count(), 4)

	//注册中心异常，重新监控
	r.exists = map[string]bool{
		"/hydra/servers": false,
	}
	r.watchErr = errors.New("error")
	r.watchChan <- &valuesEntity{values: []string{}}
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.cacheAddress.Count(), 0)
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.exists, false)

}
