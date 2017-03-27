package json

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/qxnw/hydra/conf"
)

func TestWatcher1(t *testing.T) {
	watcher := NewJSONConfWatcher()
	checker := &testfileChecker{modTime: time.Now(), apis: map[string]string{
		"../api": "../api",
	}, files: map[string]string{
		"../api/merchant.api/conf/conf.json": "../api/merchant.api/conf/conf.json",
	}}
	watcher.checker = checker
	watcher.timeSpan = time.Millisecond * 100
	f, err := watcher.Notify()
	watcher.Start()
	if err != nil {
		t.Error(err)
	}

	select {
	case updater := <-f:
		expect(t, updater.Op, conf.ADD)
		expect(t, updater.Conf.String("name"), "merchant.api")
	}
	checker.modTime = time.Now()
	select {
	case updater := <-f:
		expect(t, updater.Op, conf.CHANGE)
		expect(t, updater.Conf.String("name"), "merchant.api")
	}
	checker.files = map[string]string{}
	checker.modTime = time.Now().Add(time.Second)
	select {
	case updater := <-f:
		expect(t, updater.Op, conf.DEL)
		expect(t, updater.Conf.String("name"), "merchant.api")
	}

}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type testfileChecker struct {
	modTime time.Time
	apis    map[string]string
	files   map[string]string
}

func (f testfileChecker) Exists(filename string) bool {
	if _, ok := f.apis[filename]; ok {
		return ok
	}
	return false
}
func (f testfileChecker) LastModeTime(path string) (t time.Time, err error) {
	return f.modTime, nil
}
func (f testfileChecker) ReadDir(path string) (r []string, err error) {
	return []string{"merchant.api"}, nil
}
func (f testfileChecker) ReadAll(path string) (buf []byte, err error) {
	if _, ok := f.files[path]; ok {
		return []byte(str), nil
	}
	return nil, errors.New("file not exists")
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
