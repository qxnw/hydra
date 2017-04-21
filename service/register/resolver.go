package register

import "github.com/qxnw/hydra/service"

//registerResolver 注册中心解析器
type registerResolver struct {
}

//Resolve 从服务器获取数据
func (j *registerResolver) Resolve(adapter string, domain string, serverName string, args ...string) (c service.IServiceRegistry, err error) {
	r, err := service.GetRegistry(adapter, args...)
	if err != nil {
		return
	}
	c = newServiceRegister(domain, serverName, r)
	return
}

func init() {
	service.Register("zk", &registerResolver{})
}
