package engines

import (
	"fmt"
	"os"
	"plugin"
	"reflect"
	"sync"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/lib4go/file"
)

var components = make(map[string]func(component.IContainer) (component.IComponent, error))
var mu sync.Mutex

func getComponent(p string, e component.IContainer) (f func(component.IContainer) (component.IComponent, error), err error) {
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
	components[p] = wkr
	return components[p], nil

}
func loadComponent(path string, wkr func(component.IContainer) (component.IComponent, error), e component.IContainer) (r component.IComponent, err error) {
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
	return rwrk, nil
}

//LoadComponents 加载所有插件
func (r *ServiceEngine) LoadComponents(files ...string) error {
	for _, file := range files {
		//根据加载的文件名，获取组件
		comp, err := getComponent(file, r)
		if err != nil {
			return err
		}
		if comp == nil || reflect.ValueOf(comp).IsNil() {
			continue
		}

		//加载组件
		cmp, err := loadComponent(file, comp, r)
		if err != nil {
			return err
		}
		if cmp == nil || reflect.ValueOf(cmp).IsNil() {
			continue
		}
		services := cmp.GetGroupServices(component.GetGroupName(r.GetServerType()))
		groupName := component.GetGroupName(r.GetServerType())
		r.logger.Infof("加载组件:%s[%d] %v", file, len(services), services)
		for _, srv := range services {
			tags := cmp.GetTags(srv)
			if len(tags) == 0 {
				tags = []string{"go"}
			}
			if hander := r.GetCachedHandler(groupName, srv); hander != nil {
				r.AddCustomerTagsService(srv, hander, tags, groupName)
			}
		}
		r.AddFallbackHandlers(cmp.GetFallbackHandlers())
	}
	return nil
}
