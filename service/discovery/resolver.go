package discovery

import "github.com/qxnw/hydra/service"

//discoveryResolver 注册中心解析器
type discoveryResolver struct {
}

//Resolve 从服务器获取数据
func (j *discoveryResolver) Resolve(adapter string, domain string, tag string, args ...string) (c service.IDiscovery, err error) {
	r, err := service.GetRegistry(adapter, args...)
	if err != nil {
		return
	}
	c = newServiceDiscovery(domain, tag, r)
	return
}

func init() {
	service.Discovery("zk", &discoveryResolver{})
}
