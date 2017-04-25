package discovery

import (
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/service"
	"github.com/qxnw/lib4go/logger"
)

//discoveryResolver 注册中心解析器
type discoveryResolver struct {
}

//Resolve 从服务器获取数据
func (j *discoveryResolver) Resolve(adapter string, domain string, tag string, log *logger.Logger, servers []string) (c service.IDiscovery, err error) {
	r, err := registry.GetRegistry(adapter, log, servers)
	if err != nil {
		return
	}
	c = newServiceDiscovery(domain, tag, r)
	return
}

func init() {
	service.Discovery("zk", &discoveryResolver{})
}
