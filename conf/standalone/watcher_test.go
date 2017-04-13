package standalone

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"fmt"

	"github.com/qxnw/hydra/registry"
)

func TestWatcher1(t *testing.T) {
	watcher := NewJSONConfWatcher("/hydra", "")
	checker := &testfileChecker{modTime: time.Now(), apis: map[string]bool{
		"/hydra/servers":                             true,
		"/hydra/servers/merchant/api/conf/conf.json": true,
	}, files: map[string]string{
		"/hydra/servers/merchant/api/conf/conf.json": "/hydra/servers/merchant/api/conf/conf.json",
	}}
	watcher.checker = checker
	watcher.timeSpan = time.Millisecond * 100
	f := watcher.Notify()
	watcher.Start()
	fmt.Println("wait change!")
	updater := <-f
	fmt.Println("has change!")

	expect(t, updater.Op, registry.ADD)
	expect(t, len(f), 0)
	expect(t, updater.Conf.String("name"), "merchant.api")
	checker.modTime = time.Now().Add(time.Second)
	updater = <-f

	expect(t, updater.Op, registry.CHANGE)
	expect(t, len(f), 0)
	expect(t, updater.Conf.String("name"), "merchant.api")
	watcher.checker = &testfileChecker{}
	updater = <-f
	expect(t, updater.Op, registry.DEL)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type testfileChecker struct {
	modTime time.Time
	apis    map[string]bool
	files   map[string]string
}

func (f testfileChecker) Exists(filename string) bool {
	if v, ok := f.apis[filename]; ok {
		return v
	}
	return false
}
func (f testfileChecker) LastModeTime(path string) (t time.Time, err error) {
	return f.modTime, nil
}
func (f testfileChecker) ReadDir(path string) (r []string, err error) {
	return []string{"merchant"}, nil
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
