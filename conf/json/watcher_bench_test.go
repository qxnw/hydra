package json

import (
	"reflect"
	"testing"
	"time"

	"github.com/qxnw/hydra/conf"
)

func BenchmarkItems(t *testing.B) {
	for i := 0; i < t.N; i++ {
		watcher := NewJSONConfWatcher()
		checker := &testfileChecker{modTime: time.Now(), apis: map[string]string{
			"../api": "../api",
			"../api/merchant.api/conf/conf.json": "../api/merchant.api/conf/conf.json",
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

		updater := <-f
		expectB(t, updater.Op, conf.ADD)
		expectB(t, len(f), 0)
		expectB(t, updater.Conf.String("name"), "merchant.api")
		checker.modTime = time.Now().Add(time.Second)
		updater = <-f

		expectB(t, updater.Op, conf.CHANGE)
		expectB(t, len(f), 0)
		expectB(t, updater.Conf.String("name"), "merchant.api")
		watcher.checker = &testfileChecker{}
		updater = <-f
		expectB(t, updater.Op, conf.DEL)

	}
}
func expectB(t *testing.B, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
