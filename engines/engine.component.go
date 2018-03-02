package engines

import (
	"fmt"
	"os"
	"plugin"
	"reflect"
	"sync"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/file"
)

var components = make(map[string]component.IComponent)
var mu sync.Mutex

func loadComponent(p string, e component.IContainer) (r component.IComponent, err error) {
	path, err := file.GetAbs(p)
	if err != nil {
		return
	}
	if p, ok := components[path]; ok {
		return p, nil
	}
	mu.Lock()
	defer mu.Unlock()
	if p, ok := components[path]; ok {
		return p, nil
	}
	if _, err = os.Lstat(path); err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	pg, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("加载引擎插件失败:%s,err:%v", path, err)
	}
	work, err := pg.Lookup("GetComponent")
	if err != nil {
		return nil, fmt.Errorf("加载引擎插件%s失败未找到函数GetComponent,err:%v", path, err)
	}
	wkr, ok := work.(func(component.IContainer) (component.IComponent, error))
	if !ok {
		return nil, fmt.Errorf("加载引擎插件%s失败 GetComponent函数必须为 func() component.IComponent类型", path)
	}
	rwrk, err := wkr(e)
	if err != nil {
		return nil, fmt.Errorf("获取组件(%s)初始化失败,err:%v", path, err)
	}
	if rwrk == nil || reflect.ValueOf(rwrk).IsNil() {
		return nil, fmt.Errorf("组件(%s)为空,Component:nil", path)
	}
	err = rwrk.LoadServices()
	if err != nil {
		return nil, fmt.Errorf("组件(%s)初始化服务失败,err:%v", path, err)
	}
	components[p] = rwrk
	return components[p], nil
}

func handler(f component.IComponent) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		if r, err := f.Handling(name, mode, service, ctx); err != nil {
			return r, err
		}
		rx, err := f.Handle(name, mode, service, ctx)
		if err != nil {
			return rx, err
		}
		if r, err := f.Handled(name, mode, service, ctx); err != nil {
			return r, err
		}
		return rx, nil
	}
}

//LoadComponents 加载所有插件
func (r *ServiceEngine) LoadComponents(files ...string) error {
	for _, file := range files {
		cmp, err := loadComponent(file, r)
		if err != nil {
			return err
		}
		if cmp == nil || reflect.ValueOf(cmp).IsNil() {
			continue
		}
		service := cmp.GetGroupServices(component.GetGroupName(r.serverType))
		r.logger.Infof("加载组件:%s[%d]", file, len(service))
		for _, srvs := range service {
			tags := cmp.GetTags(srvs)
			if len(tags) == 0 {
				tags = []string{"go"}
			}
			r.AddCustomerTagsService(srvs, handler(cmp), tags, component.GetGroupName(r.serverType))
			r.logger.Debug("加载外部服务:", srvs,)
		}
		
	}
	return nil
}
