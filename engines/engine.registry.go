package engines

import (
	"fmt"

	"github.com/qxnw/hydra/component"
)

var serviceLoader = make(map[string]func(r *component.StandardComponent, i component.IContainer))

//AddServiceLoader 添加服务加载器
func AddServiceLoader(name string, f func(r *component.StandardComponent, i component.IContainer)) {
	if _, ok := serviceLoader[name]; ok {
		panic(fmt.Sprintf("重复注册服务:%s", name))
	}
	serviceLoader[name] = f

}

func (r *ServiceEngine) loadEngineServices() {
	for _, engine := range r.engines {
		if loader, ok := serviceLoader[engine]; ok {
			loader(r.StandardComponent, r)
		} else {
			fmt.Sprintln("未能加载引擎:", engine)
		}
	}
}
