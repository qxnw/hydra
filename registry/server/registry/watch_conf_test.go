package registry

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/conf/server"
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
	watchChan chan registry.ValueWatcher
	watchErr  error
	value     []byte
	exists    map[string]bool
}

func (r *confRegistry) Exists(path string) (bool, error) {
	if v, ok := r.exists[path]; ok {
		return v, nil
	}
	return false, nil
}
func (r *confRegistry) WatchValue(path string) (data chan registry.ValueWatcher, err error) {
	data = r.watchChan
	err = r.watchErr
	return
}

func (r *confRegistry) GetValue(path string) (data []byte, err error) {
	data = r.value
	return
}

func (r *confRegistry) WatchChildren(path string) (data chan registry.ChildrenWatcher, err error) {
	return nil, nil
}
func (r *confRegistry) GetChildren(path string) (data []string, err error) {
	return nil, nil
}

func TestConfWacher2(t *testing.T) {
	r := &confRegistry{
		watchChan: make(chan registry.ValueWatcher, 1),
		value:     []byte(str),
	}
	updater := make(chan *server.Updater, 1)
	watcher := NewWatchConf("/hydra/servers/merchant/api/conf/192.168.0.100:01", r, updater, time.Millisecond*100)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": true,
	}
	r.value = []byte(str)
	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.add")
	case v := <-updater:
		expect(t, v.Op, conf.ADD)
		expect(t, len(updater), 0)
		expect(t, v.Conf.String("name"), "merchant.api")
	}
	watcher.Close()
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}
}

func TestConfWacher1(t *testing.T) {
	r := &confRegistry{
		watchChan: make(chan registry.ValueWatcher, 1),
		value:     []byte(str),
	}
	updater := make(chan *server.Updater, 1)
	watcher := NewWatchConf("/hydra/servers/merchant/api/conf/192.168.0.100:01", r, updater, time.Millisecond*100)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": true,
	}
	r.value = []byte(str)
	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.add")
	case v := <-updater:
		expect(t, v.Op, conf.ADD)
		expect(t, len(updater), 0)
		expect(t, v.Conf.String("name"), "merchant.api")
	}

	//节点发生变化
	r.watchChan <- &valueEntity{Value: []byte(str)}
	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.change")
	case v := <-updater:
		expect(t, v.Op, conf.CHANGE)
		expect(t, len(updater), 0)
		expect(t, v.Conf.String("name"), "merchant.api")
	}
	//节点再次发生变化
	r.watchChan <- &valueEntity{Value: []byte(str)}
	select {
	case <-time.After(time.Millisecond * 200):
		expect(t, nil, "conf.CHANGE")
	case v := <-updater:
		expect(t, v.Op, conf.CHANGE)
		expect(t, len(updater), 0)
		expect(t, v.Conf.String("name"), "merchant.api")
	}
	//与注册中心交互发生异常1:获取的值为空
	r.value = []byte("")
	r.watchChan <- &valueEntity{Value: []byte("")}
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}
	//与注册中心交互发生异常1:获取的值为空，节点不存在
	r.value = []byte("")
	r.exists = map[string]bool{
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": false,
	}
	r.watchErr = errors.New("error")
	r.watchChan <- &valueEntity{Value: []byte("")}
	select {
	case v := <-updater:
		expect(t, v, nil)
	default:
	}

}

func TestConfWacher3(t *testing.T) {
	r := &confRegistry{
		watchChan: make(chan registry.ValueWatcher, 1),
		value:     []byte(str),
	}
	updater := make(chan *server.Updater, 1)
	watcher := NewWatchConf("/hydra/servers/merchant/api/conf/192.168.0.100:01", r, updater, time.Millisecond*100)
	go watcher.watch()

	//正常已存在的节点
	r.exists = map[string]bool{
		"/hydra/servers/merchant/api/conf/192.168.0.100:01": true,
	}
	r.value = []byte(str)
	select {
	case v := <-updater:
		expect(t, v.Op, conf.ADD)
		expect(t, len(updater), 0)
		expect(t, v.Conf.String("name"), "merchant.api")
	}
	watcher.NotifyConfDel()
	select {
	case v := <-updater:
		expect(t, v.Op, conf.DEL)
	default:
	}
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

var str = `{
    "type": "api",
    "name": "merchant.api",
    "status": "starting",
    "package": "1.0.0.1",
	"root":    "@type/@name",
    "QPS": 1000,
    "limit": [
        {
            "local": "@client like 192.168*",
            "Operation": "deny",
            "to": "@service like 192.168*"
        },
        {
            "local": "@client like 192.168* && @service == /order/request",
            "Operation": "assign",
            "to": "@ip like 192.168.1*"
        }
    ],
    "routes": [
        {
            "from": "/:module/:action/:id",
            "method": "request",
            "to": "../@type/@name/script/@module_@action:@method",
            "params": "db=@var_weixin"
        }
    ]
}`

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
