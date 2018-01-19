package plugin

import (
	"fmt"
	"os"
	"plugin"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/file"
)

func (s *goPluginWorker) loadComponent(p string) (r *workerService, err error) {
	mu.Lock()
	defer mu.Unlock()
	path, err := file.GetAbs(p)
	if err != nil {
		return
	}
	if p, ok := plugines[path]; ok {
		return p, nil
	}
	if _, err = os.Lstat(p); err != nil && os.IsNotExist(err) {
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
	wkr, ok := work.(func(engine.IContainer)(component.IComponent,error))
	if !ok{
		return nil, fmt.Errorf("加载引擎插件%s失败 GetComponent函数必须为 func() component.IComponent类型", path)
	}
	context := engine.NewContainer(s.ctx.RPC, s.ctx.Registry, s.ctx.Domain, s.ctx.ServerName, s.ctx.Logger)
	rwrk,err := wkr(context)
	if err != nil {
		return nil, fmt.Errorf("获取组件(%s)初始化失败,err:%v", path, err)
	}
	err = rwrk.LoadServices()
	if err != nil {
		return nil, fmt.Errorf("组件(%s)初始化服务失败,err:%v", path, err)
	}
	srvs := rwrk.GetServices()
	plugines[path] = &workerService{IComponent: rwrk, services: srvs}
	s.ctx.Logger.Info("加载引擎插件：", p)
	return plugines[path], nil
}
