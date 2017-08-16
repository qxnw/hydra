package plugin

import (
	"fmt"
	"os"
	"plugin"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/file"
)

func (s *goPluginWorker) loadPlugin(p string) (r context.Worker, err error) {
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
	work, err := pg.Lookup("GetWorker")
	if err != nil {
		return nil, fmt.Errorf("加载引擎插件%s失败未找到函数GetWorker,err:%v", path, err)
	}
	wkr, ok := work.(func() context.Worker)
	if !ok {
		return nil, fmt.Errorf("加载引擎插件%s失败 GetWorker函数必须为 func() context.Worker类型", path)
	}
	rwrk := wkr()
	plugines[path] = rwrk
	s.ctx.Logger.Info("加载引擎插件：", p)
	return rwrk, nil
}
