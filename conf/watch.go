package conf

import (
	"fmt"

	"github.com/qxnw/lib4go/logger"
)

//WatchServers /api，mq，rpc，cron
var WatchServers = []string{"api", "rpc", "cron", "mqc", "web"}

type Updater struct {
	Conf Conf
	Op   int
}

//Watcher 配置文件监控器
type Watcher interface {
	Start() error
	Notify() chan *Updater
	Close() error
}

//WatcherResolver 定义配置文件转换方法
type WatcherResolver interface {
	Resolve(adapter string, domain string, tag string, log *logger.Logger, servers []string) (Watcher, error)
}

var confResolvers = make(map[string]WatcherResolver)

//Register 注册配置文件适配器
func Register(name string, resolver WatcherResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := confResolvers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	confResolvers[name] = resolver
}

//NewWatcher 根据适配器名称及参数返回配置处理器
func NewWatcher(adapter string, domain string, tag string, log *logger.Logger, servers []string) (Watcher, error) {
	resolver, ok := confResolvers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, tag, log, servers)
}
