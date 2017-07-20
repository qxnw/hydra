package cluster

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//RegistryConfResolver 注册中心解析器
type RegistryConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *RegistryConfResolver) Resolve(adapter string, domain string, confTag string, log *logger.Logger, servers []string) (c conf.Watcher, err error) {
	r, err := registry.NewRegistry(adapter, servers, log)
	if err != nil {
		return
	}
	c = newRegistryConfWatcher(domain, confTag, r, log)
	return
}

func init() {
	conf.Register("zk", &RegistryConfResolver{})
}
