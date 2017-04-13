package discovery

import "github.com/qxnw/lib4go/registry"

type watcherRegistry struct {
	watchChilrenChan chan registry.ChildrenWatcher
	watchValueChan   chan registry.ValueWatcher
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
func (r *watcherRegistry) WatchValue(path string) (data chan registry.ValueWatcher, err error) {
	data = r.watchValueChan
	return
}

func (r *watcherRegistry) GetValue(path string) (data []byte, err error) {
	return r.value, nil
}

func (r *watcherRegistry) WatchChildren(path string) (data chan registry.ChildrenWatcher, err error) {
	data = r.watchChilrenChan
	err = r.watchErr
	return
}
func (r *watcherRegistry) GetChildren(path string) (data []string, err error) {
	data = r.children
	return
}

/*
func TestWatcher1(t *testing.T) {
	r := &watcherRegistry{
		watchValueChan:   make(chan registry.ValueWatcher, 1),
		watchChilrenChan: make(chan registry.ChildrenWatcher, 1),
	}
	r.exists = map[string]bool{
		"/hydra/services":                                    true,
		"/hydra/services/merchant.api/order.query/providers": true,
	}
	watcher := NewRegistryConfWatcher("/hydra", "merchant.api", r)
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

	r.watchChilrenChan <- &valuesEntity{values: []string{"order.query"}}
	time.Sleep(time.Millisecond * 100)
	wp, ok := watcher.watchPaths.Get("/hydra/services")
	expect(t, ok, true)
	if ok {
		expect(t, wp.(*watchPath).cacheAddress.Count(), 1)
	}
	select {
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, conf.ADD)
	default:
	}

	r.watchChilrenChan <- &valuesEntity{values: []string{}}
	select {
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, conf.DEL)
	default:
	}
	expect(t, len(r.watchChilrenChan), 0)
	r.watchChilrenChan <- &valuesEntity{values: []string{"merchant"}}
	select {
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, conf.ADD)
	default:
	}
	time.Sleep(time.Millisecond * 10)
	r.exists = map[string]bool{
		"/hydra/services":                                    false,
		"/hydra/services/merchant.api/order.query/providers": false,
	}
	r.watchErr = errors.New("error")
	expect(t, len(r.watchChilrenChan), 0)
	r.watchChilrenChan <- &valuesEntity{values: []string{"merchant.api"}}
	select {
	case v := <-updater:
		expect(t, len(v), 1)
		expect(t, v[0].Op, conf.DEL)
	default:
	}
}
*/
