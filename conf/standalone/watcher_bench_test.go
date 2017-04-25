package standalone

import (
	"reflect"
	"testing"
	"time"

	"github.com/qxnw/hydra/registry"
)

func BenchmarkItems(t *testing.B) {
	for i := 0; i < t.N; i++ {
		watcher := NewJSONConfWatcher("/hydra", "")
		checker := &testfileChecker{modTime: time.Now(), apis: map[string]bool{
			"/hydra/servers":                        true,
			"/hydra/servers/merchant/api/conf/conf": true,
		}, files: map[string]string{
			"/hydra/servers/merchant/api/conf/conf": "/hydra/servers/merchant/api/conf/conf",
		}}
		watcher.checker = checker
		watcher.timeSpan = time.Millisecond * 100
		f := watcher.Notify()
		watcher.Start()

		updater := <-f
		expectB(t, updater.Op, registry.ADD)
		expectB(t, len(f), 0)
		expectB(t, updater.Conf.String("name"), "merchant.api")
		checker.modTime = time.Now().Add(time.Second)
		updater = <-f

		expectB(t, updater.Op, registry.CHANGE)
		expectB(t, len(f), 0)
		expectB(t, updater.Conf.String("name"), "merchant.api")
		watcher.checker = &testfileChecker{}
		updater = <-f
		expectB(t, updater.Op, registry.DEL)

	}
}
func expectB(t *testing.B, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}