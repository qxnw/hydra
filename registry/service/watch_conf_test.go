package service

import (
	"reflect"
	"testing"
	"time"

	re "github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/registry"
)

type kv struct {
	value string
	err   error
}
type ckv struct {
	ch  chan []string
	err error
}

type confRegistry struct {
	childrenChan chan registry.ChildrenWatcher
	children     []string
	watchErr     error
	exists       map[string]bool
}

func (r *confRegistry) Exists(path string) (bool, error) {
	if v, ok := r.exists[path]; ok {
		return v, nil
	}
	return false, nil
}
func (r *confRegistry) WatchValue(path string) (data chan registry.ValueWatcher, err error) {
	return nil, nil
}

func (r *confRegistry) GetValue(path string) (data []byte, err error) {
	return nil, nil
}

func (r *confRegistry) WatchChildren(path string) (data chan registry.ChildrenWatcher, err error) {
	return r.childrenChan, nil
}
func (r *confRegistry) GetChildren(path string) (data []string, err error) {
	return r.children, nil
}
func (r *confRegistry) CreatePersistentNode(path string, data string) (err error) {
	return nil
}
func (r *confRegistry) CreateTempNode(path string, data string) (err error) {
	return nil
}
func (r *confRegistry) CreateSeqNode(path string, data string) (rpath string, err error) {
	return "", nil
}
func TestConfWacher2(t *testing.T) {
	r := &confRegistry{
		childrenChan: make(chan registry.ChildrenWatcher, 1),
	}
	updater := make(chan []*ServiceUpdater, 1)
	watcher := NewWatchConf("/hydra/services/merchant.api/order.request/providers", r, updater, time.Millisecond*100)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/services/merchant.api/order.request/providers": true,
	}
	r.childrenChan <- &valuesEntity{values: []string{"192.168.0.1"}}

	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.add")
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, re.ADD)
	}

	//再加添加不同机器
	r.childrenChan <- &valuesEntity{values: []string{"192.168.0.2", "192.168.0.1"}}
	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.add")
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, re.ADD)
	}
	watcher.Close()
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}

}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type valueEntity struct {
	Value []byte
	Err   error
}
type valuesEntity struct {
	values []string
	Err    error
}

func (v *valueEntity) GetValue() []byte {
	return v.Value
}
func (v *valueEntity) GetError() error {
	return v.Err
}

func (v *valuesEntity) GetValue() []string {
	return v.values
}
func (v *valuesEntity) GetError() error {
	return v.Err
}
