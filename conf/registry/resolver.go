package registry

import (
	"errors"

	"github.com/qxnw/hydra/conf"
)

//ZookeeperConfResolver zookeeper配置引擎
type RegistryConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *RegistryConfResolver) Resolve(registry Registry, args ...string) (conf.ConfWatcher, error) {
	if len(args) < 3 {
		return nil, errors.New("输入参数不能为空")
	}
	domain := args[1]
	tag := args[2]
	return NewRegistryConfWatcher(domain, tag, registry), nil
}

/*
func init() {
	conf.Register("zk", &RegistryConfResolver{})
}
*/
