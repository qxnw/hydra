package registry

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"strings"

	"github.com/qxnw/hydra/conf"
)

type registry struct {
	children chan []string
	values   chan string
	value    string
	exists   bool
}

func (r *registry) Exists(path string) bool {
	return r.exists
}
func (r *registry) WatchChildren(path string) (data chan []string, err error) {
	data = make(chan []string, 1)
	if strings.HasSuffix(path, "api") {
		data <- <-r.children
		return
	}
	<-data
	return
}
func (r *registry) WatchValue(path string) (data chan string, err error) {
	data = make(chan string, 1)
	data <- <-r.values
	return
}
func (r *registry) GetChildren(path string) (data []string, err error) {
	data = make([]string, 0, 0)
	return
}
func (r *registry) GetValue(path string) (data string, err error) {
	data = r.value
	return
}

func TestWatcher1(t *testing.T) {
	r := &registry{
		children: make(chan []string, 1),
		values:   make(chan string, 2),
		value:    str,
	}

	watcher := NewRegistryConfWatcher("/hydra", "192.168.0.1", r)
	watcher.timeSpan = time.Millisecond * 100
	f, err := watcher.Notify()
	watcher.Start()
	if err != nil {
		t.Error(err)
	}
	r.children <- []string{"merchant.api"}
	r.exists = true

	updater := <-f
	expect(t, updater.Op, conf.ADD)
	expect(t, len(f), 0)
	expect(t, updater.Conf.String("name"), "merchant.api")

	r.values <- str
	updater = <-f
	expect(t, updater.Op, conf.CHANGE)
	expect(t, len(f), 0)
	expect(t, updater.Conf.String("name"), "merchant.api")

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
