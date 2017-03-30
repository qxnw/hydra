package registry

import "github.com/qxnw/lib4go/registry"

//Registry 注册中心接口
type Registry interface {
	Exists(path string) (bool, error)
	WatchChildren(path string) (data chan registry.ChildrenWatcher, err error)
	WatchValue(path string) (data chan registry.ValueWatcher, err error)
	GetChildren(path string) (data []string, err error)
	GetValue(path string) (data []byte, err error)
}
