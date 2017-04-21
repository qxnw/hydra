package registry

import "github.com/qxnw/lib4go/registry"

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
