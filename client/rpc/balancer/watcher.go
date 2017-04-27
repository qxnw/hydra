package balancer

import (
	"errors"

	"strings"

	"github.com/qxnw/hydra/registry"
	r "github.com/qxnw/lib4go/registry"

	"sort"

	"time"

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
}

// Close do nothing
func (w *Watcher) Close() {
	close(w.closeCh)
}

// Next to return the updates
func (w *Watcher) Next() ([]*naming.Update, error) {
	updates := make([]*naming.Update, 0, 4)
	if !w.isInitialized {
		resp, _, err := w.client.GetChildren(w.service)
		w.isInitialized = true
		if err == nil {
			addrs := w.extractAddrs(resp)
			if l := len(addrs); l != 0 {
				for i := range addrs {
					updates = append(updates, &naming.Update{Op: naming.Add, Addr: addrs[i]})
				}
				return updates, nil
			}
		}
	}

	// generate etcd Watcher
	watcherCh, err := w.client.WatchChildren(w.service)
	if err != nil {
		time.Sleep(time.Second)
		return nil, nil
	}
	var watcher r.ChildrenWatcher
	select {
	case watcher = <-watcherCh:
	case <-w.closeCh:
		return nil, errors.New("watcher closed")
	}
	if err = watcher.GetError(); err != nil {
		time.Sleep(time.Second)
		return nil, nil
	}
	chilren, _ := watcher.GetValue()
	addrs := w.extractAddrs(chilren)
	newCache := make(map[string]bool)
	for i := 0; i < len(addrs); i++ {
		if _, ok := w.caches[addrs[i]]; !ok {
			updates = append(updates, &naming.Update{Op: naming.Add, Addr: addrs[i]})
			newCache[addrs[i]] = true
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
	return updates, nil
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
