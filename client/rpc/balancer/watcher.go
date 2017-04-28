package balancer

import (
	"strings"

	"github.com/qxnw/hydra/registry"
	r "github.com/qxnw/lib4go/registry"

	"sort"

	"google.golang.org/grpc/naming"
)

// Watcher is the implementaion of grpc.naming.Watcher
type Watcher struct {
	client        registry.Registry
	isInitialized bool
	caches        map[string]bool
	service       string
	sortPrefix    string
	closeCh       chan struct{}
	lastErr       error
}

// Close do nothing
func (w *Watcher) Close() {
	close(w.closeCh)
}

// Next to return the updates
func (w *Watcher) Next() ([]*naming.Update, error) {
	w.lastErr = nil
	if !w.isInitialized {
		resp, _, err := w.client.GetChildren(w.service)
		w.isInitialized = true
		if err == nil {
			addrs := w.extractAddrs(resp)
			return w.getUpdates(addrs), nil
		}
	}

	// generate etcd Watcher
	watcherCh, err := w.client.WatchChildren(w.service)
	if err != nil {
		return w.getUpdates([]string{}), err
	}
	var watcher r.ChildrenWatcher
	select {
	case watcher = <-watcherCh:
	case <-w.closeCh:
		return w.getUpdates([]string{}), w.lastErr
	}
	if err = watcher.GetError(); err != nil {
		return w.getUpdates([]string{}), err
	}
	chilren, _ := watcher.GetValue()
	addrs := w.extractAddrs(chilren)
	return w.getUpdates(addrs), nil
}
func (w *Watcher) getUpdates(addrs []string) (updates []*naming.Update) {
	newCache := make(map[string]bool)
	for i := 0; i < len(addrs); i++ {
		newCache[addrs[i]] = true
		if _, ok := w.caches[addrs[i]]; !ok {
			updates = append(updates, &naming.Update{Op: naming.Add, Addr: addrs[i]})
		} else {
			w.caches[addrs[i]] = false
		}
	}
	for i, v := range w.caches {
		if v {
			updates = append(updates, &naming.Update{Op: naming.Delete, Addr: i})
		}
	}
	w.caches = newCache
	return
}
func (w *Watcher) extractAddrs(resp []string) []string {
	addrs := make([]string, 0, len(resp))
	for _, v := range resp {
		item := strings.SplitN(v, "_", 2)
		addrs = append(addrs, item[0])
	}
	if w.sortPrefix != "" {
		sort.Slice(addrs, func(i, j int) bool {
			return strings.HasPrefix(addrs[i], w.sortPrefix)
		})
	}
	return addrs
}
