package cluster

import (
	"errors"
	"testing"
	"time"

	"github.com/qxnw/hydra/registry"
	rx "github.com/qxnw/lib4go/registry"
)

type watcherRegistry struct {
	watchChilrenChan chan rx.ChildrenWatcher
	watchValueChan   chan rx.ValueWatcher
	watchErr         error
	exists           map[string]bool
	children         []string
	value            []byte
}

func (r *watcherRegistry) Exists(path string) (bool, error) {
	if v, ok := r.exists[path]; ok {
		return v, nil
	}
	return false, nil
}
func (r *watcherRegistry) WatchValue(path string) (data chan rx.ValueWatcher, err error) {
	data = r.watchValueChan
	return
}

func (r *watcherRegistry) GetValue(path string) (data []byte, err error) {
	return r.value, nil
}

func (r *watcherRegistry) WatchChildren(path string) (data chan rx.ChildrenWatcher, err error) {
	data = r.watchChilrenChan
	err = r.watchErr
	return
}
func (r *watcherRegistry) GetChildren(path string) (data []string, err error) {
	data = r.children
	return
}
func (r *watcherRegistry) CreatePersistentNode(path string, data string) (err error) {
	return nil
}
func (r *watcherRegistry) CreateTempNode(path string, data string) (err error) {
	return nil
}
func (r *watcherRegistry) CreateSeqNode(path string, data string) (rpath string, err error) {
	return "", nil
}
func TestWatcher1(t *testing.T) {
	r := &watcherRegistry{
		watchValueChan:   make(chan rx.ValueWatcher, 1),
		watchChilrenChan: make(chan rx.ChildrenWatcher, 1),
	}
	r.exists = map[string]bool{
		"/hydra/servers":                                    true,
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": true,
	}
	watcher := NewRegistryConfWatcher("/hydra", "192.168.0.100:01", r)
	watcher.timeSpan = time.Millisecond * 10
	watcher.Start()
	time.Sleep(time.Millisecond * 10)
	expect(t, watcher.watchPaths.Count(), 1)

	updater := watcher.Notify()

	//正常已存在的节点
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}

	r.watchChilrenChan <- &valuesEntity{values: []string{"merchant"}}
	time.Sleep(time.Millisecond * 100)
	wp, ok := watcher.watchPaths.Get("/hydra/servers")
	expect(t, ok, true)
	if ok {
		expect(t, wp.(*watchPath).cacheAddress.Count(), 4)
	}
	select {
	case v := <-updater:
		expect(t, v.Op, registry.ADD)
	default:
	}

	r.watchChilrenChan <- &valuesEntity{values: []string{}}
	select {
	case v := <-updater:
		expect(t, v.Op, registry.DEL)
	default:
	}
	expect(t, len(r.watchChilrenChan), 0)
	r.watchChilrenChan <- &valuesEntity{values: []string{"merchant"}}
	select {
	case v := <-updater:
		expect(t, v.Op, registry.ADD)
	default:
	}
	time.Sleep(time.Millisecond * 10)
	r.exists = map[string]bool{
		"/hydra/servers":                                    false,
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": false,
	}
	r.watchErr = errors.New("error")
	expect(t, len(r.watchChilrenChan), 0)
	r.watchChilrenChan <- &valuesEntity{values: []string{"merchant"}}
	select {
	case v := <-updater:
		expect(t, v.Op, registry.DEL)
	default:
	}
}
