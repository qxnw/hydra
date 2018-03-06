package engines

import (
	"fmt"

	"github.com/qxnw/hydra/component"
)

//ServiceLoader 服务加载器
type ServiceLoader func(r *component.StandardComponent, i component.IContainer)

var serviceLoaders = make(map[string]ServiceLoader)

//AddGoLoader 添加go引擎的服务加载器
func AddGoLoader(f ServiceLoader) {
	AddLoader("go", f)
}

//AddLoader 添加服务加载器
func AddLoader(name string, f ServiceLoader) {
	if _, ok := serviceLoaders[name]; ok {
		panic(fmt.Sprintf("重复注册服务:%s", name))
	}
	serviceLoaders[name] = f
}

func (r *ServiceEngine) loadEngineServices() {
	for _, engine := range r.engines {
		if loader, ok := serviceLoaders[engine]; ok {
			loader(r.StandardComponent, r)
		}
	}
}
