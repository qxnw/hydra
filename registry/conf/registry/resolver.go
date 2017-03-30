package registry

import "github.com/qxnw/hydra/conf/server"

//RegistryConfResolver 注册中心解析器
type RegistryConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *RegistryConfResolver) Resolve(adapter string, domain string, tag string, args ...string) (c server.ConfWatcher, err error) {
	r, err := getRegistry(adapter, args...)
	if err != nil {
		return
	}
	c = NewRegistryConfWatcher(domain, tag, r)
	return
}

func init() {
	server.Register("zookeeper", &RegistryConfResolver{})
}
