package register

import (
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/service"
	"github.com/qxnw/lib4go/logger"
)

//clusterRegisterResolver 注册中心解析器
type clusterRegisterResolver struct {
}

//Resolve 从服务器获取数据
func (j *clusterRegisterResolver) Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (c service.IServiceRegistry, err error) {
	r, err := registry.NewRegistry(adapter, servers, log)
	if err != nil {
		return
	}
	var crossRegistry registry.Registry
	if len(cross) == 0 {
		c = newClusterServiceRegister(domain, serverName, r, crossRegistry)
		return
	}
	crossRegistry, err = registry.NewRegistry(adapter, cross, log)
	if err != nil {
		return
	}
	c = newClusterServiceRegister(domain, serverName, r, crossRegistry)
	return
}

//standaloneRegisterResolver 注册中心解析器
type standaloneRegisterResolver struct {
}

//Resolve 从服务器获取数据
func (j *standaloneRegisterResolver) Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string, cross []string) (c service.IServiceRegistry, err error) {
	r, err := registry.NewChecker()
	if err != nil {
		return
	}
	c = newStandaloneServiceRegister(domain, serverName, r)
	return
}

func init() {
	service.Register("zk", &clusterRegisterResolver{})
	service.Register("standalone", &standaloneRegisterResolver{})
}
