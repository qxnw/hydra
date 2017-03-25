package zookeeper

import (
	"errors"
	"fmt"

	"strings"

	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/zk"
)

//ZookeeperConfAdapter zookeeper配置引擎
type ZookeeperConfResolver struct {
}
type zookeeperConfWatcher struct {
	client        *zk.Client
	root          string
	isInitialized bool
}

//Parse 从服务器获取数据
func (j *ZookeeperConfResolver) Resolve(key ...string) (ConfWatcher, error) {
	if len(args) < 2 {
		return nil, errors.New("输入参数不能为空")
	}
	servers := args[0]
	root := args[1]
	client, err := zk.New(strings.Split(servers, ";"), time.Second*3)
	if err != nil {
		return
	}
	return &ZookeeperConfWatcher{client: client, root: root}
}
func (w *zookeeperConfWatcher) Next() (chan []Updater, error) {
	// prefix is the etcd prefix/value to watch
	servicePath := fmt.Sprintf("%s", root)
	updates := make([]*conf.Updater, 0, 4)
	// check if is initialized
	if !w.isInitialized {
		// query addresses from etcd
		servicePath := fmt.Sprintf("%s", root)
		resp, err := w.client.GetChildren(servicePath)
		w.isInitialized = true
		if err == nil {
			addrs := w.extractAddrs(resp)
			//if not empty, return the updates or watcher new dir
			if l := len(addrs); l != 0 {
				for i := range addrs {
					updates = append(updates, &conf.Update{Op: conf.ADD, Addr: addrs[i]})
				}
				return updates, nil
			}
		}
	}

	// generate etcd Watcher
	resp, err := w.client.WatchChildren(servicePath)
	if err != nil {
		return nil, err
	}
	addrs := w.extractAddrs(resp)
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
