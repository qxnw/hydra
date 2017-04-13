package discovery

import (
	"testing"
	"time"

	"errors"

	re "github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/registry"
)

type pathRegistry struct {
	watchChan chan registry.ChildrenWatcher
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
func (r *pathRegistry) WatchValue(path string) (data chan registry.ValueWatcher, err error) {
	return nil, nil
}

func (r *pathRegistry) GetValue(path string) (data []byte, i int32, err error) {
	return nil, 0, nil
}

func (r *pathRegistry) WatchChildren(path string) (data chan registry.ChildrenWatcher, err error) {
	data = r.watchChan
	err = r.watchErr
	return
}
func (r *pathRegistry) GetChildren(path string) (data []string, i int32, err error) {
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
func (r *pathRegistry) Delete(path string) (err error) {
	return nil
}
func TestPathWatcher1(t *testing.T) {
	r := &pathRegistry{
		watchChan: make(chan registry.ChildrenWatcher, 1),
	}
	updater := make(chan []*re.ServiceUpdater, 4)
	watcher := NewWatchPath("/hydra", "merchant.api", "/hydra/services", r, updater, time.Millisecond*10)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/services": true,
	}
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}
	r.watchChan <- &valuesEntity{values: []string{"merchant.api"}}
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.exists, true)
	expect(t, watcher.cacheAddress.Count(), 1)

	//注册中心异常，重新监控
	r.exists = map[string]bool{
		"/hydra/services": false,
	}
	r.watchErr = errors.New("error")
	r.watchChan <- &valuesEntity{values: []string{}}
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.cacheAddress.Count(), 0)
	time.Sleep(time.Millisecond * 11)
	expect(t, watcher.exists, false)

}
