package register

import (
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/service"
	"github.com/qxnw/lib4go/logger"
)

//registerResolver 注册中心解析器
type registerResolver struct {
}

//Resolve 从服务器获取数据
func (j *registerResolver) Resolve(adapter string, domain string, serverName string, log *logger.Logger, servers []string) (c service.IServiceRegistry, err error) {
	r, err := registry.NewRegistry(adapter, servers, log)
	if err != nil {
		return
	}
	c = newServiceRegister(domain, serverName, r)
	return
}

func init() {
	service.Register("zk", &registerResolver{})
}
