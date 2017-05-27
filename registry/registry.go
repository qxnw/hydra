package registry

import (
	"fmt"

	"strings"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/registry"
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
	Update(path string, data string, version int32) (err error)
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

//RegistryResolver 定义配置文件转换方法
type RegistryResolver interface {
	Resolve(servers []string, log *logger.Logger) (Registry, error)
}

var registryResolvers = make(map[string]RegistryResolver)

//Register 注册配置文件适配器
func Register(name string, resolver RegistryResolver) {
	if resolver == nil {
		panic("registry: Register adapter is nil")
	}
	if _, ok := registryResolvers[name]; ok {
		panic("registry: Register called twice for adapter " + name)
	}
	registryResolvers[name] = resolver
}

//NewRegistry 创建注册中心
func NewRegistry(name string, servers []string, log *logger.Logger) (r Registry, err error) {
	resolver, ok := registryResolvers[name]
	if !ok {
		return nil, fmt.Errorf("registry: unknown adapter name %q (forgotten import?)", name)
	}
	key := fmt.Sprintf("%s_%s", name, strings.Join(servers, "_"))
	_, value, err := registryMap.SetIfAbsentCb(key, func(input ...interface{}) (interface{}, error) {
		rsvr := input[0].(RegistryResolver)
		srvs := input[1].([]string)
		log := input[2].(*logger.Logger)
		return rsvr.Resolve(srvs, log)
	}, resolver, servers, log)
	return value.(Registry), err
}

//NewRegistryWithAddress 根据协议地址创建注册中心
func NewRegistryWithAddress(address string, log *logger.Logger) (r Registry, err error) {
	proto, addrss, err := ResolveAddress(address)
	if err != nil {
		return nil, err
	}
	return NewRegistry(proto, addrss, log)
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

//ResolveAddress 解析地址
//如:zk://192.168.0.155:2181
//如:standalone://localhost
func ResolveAddress(address string) (proto string, raddr []string, err error) {
	addr := strings.SplitN(address, "://", 2)
	if len(addr) != 2 {
		return "", nil, fmt.Errorf("%s错误，必须包含://", addr)
	}
	if len(addr[0]) == 0 {
		return "", nil, fmt.Errorf("%s错误，协议名不能为空", addr)
	}
	if len(addr[1]) == 0 {
		return "", nil, fmt.Errorf("%s错误，地址不能为空", addr)
	}
	proto = addr[0]
	raddr = strings.Split(addr[1], ",")
	return
}
