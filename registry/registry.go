package registry

import (
	"fmt"
	"time"

	"strings"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/registry"
	"github.com/qxnw/lib4go/zk"
)

//Registry 注册中心接口
type Registry interface {
	Exists(path string) (bool, error)
	WatchChildren(path string) (data chan registry.ChildrenWatcher, err error)
	WatchValue(path string) (data chan registry.ValueWatcher, err error)
	GetChildren(path string) (data []string, version int32, err error)
	GetValue(path string) (data []byte, version int32, err error)
	CreatePersistentNode(path string, data string) (err error)
	CreateTempNode(path string, data string) (err error)
	CreateSeqNode(path string, data string) (rpath string, err error)
	Delete(path string) error
	Close()
}

type ServiceUpdater struct {
	Value string
	Op    int
}

const (
	ADD = iota + 1
	CHANGE
	DEL
)

//GetRegistry 获取注册中心
var registryMap cmap.ConcurrentMap

func init() {
	registryMap = cmap.New()
}

func GetRegistry(name string, log *logger.Logger, servers []string) (r Registry, err error) {
	switch name {
	case "zk":
		key := fmt.Sprintf("%s_%s", name, strings.Join(servers, "_"))
		_, value, err := registryMap.SetIfAbsentCb(key, func(input ...interface{}) (interface{}, error) {
			zclient, err := zk.NewWithLogger(servers, time.Second, log)
			if err != nil {
				return nil, err
			}
			err = zclient.Connect()
			return zclient, err
		})
		r = value.(Registry)
		return r, err

	}
	return nil, fmt.Errorf("不支持的注册中心:%s", name)
}

//Close 关闭注册中心的服务
func Close() {
	registryMap.RemoveIterCb(func(key string, value interface{}) bool {
		if v, ok := value.(Registry); ok {
			v.Close()
		}
		return true
	})
}
