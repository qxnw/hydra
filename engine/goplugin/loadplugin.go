package goplugin

import (
	"fmt"
	"os"
	"plugin"

	"github.com/qxnw/goplugin"
	"github.com/qxnw/lib4go/file"
)

func (s *goPluginWorker) loadPlugin(p string) (r goplugin.Worker, err error) {
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
		return nil, fmt.Errorf("加载失败:%s,err:%v", path, err)
	}
	work, err := pg.Lookup("GetWorker")
	if err != nil {
		return nil, fmt.Errorf("未找到函数GetWorker:%s,err:%v", path, err)
	}
	wkr, ok := work.(func() goplugin.Worker)
	if !ok {
		return nil, fmt.Errorf("GetWorker函数必须为 func() PluginWorker 类型:%s", path)
	}
	rwrk := wkr()
	plugines[path] = rwrk
	return rwrk, nil
}
