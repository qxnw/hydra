package http

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
)

//NeedRestart 检查配置判断是否需要重启服务器
func (w *WebResponsiveServer) NeedRestart(cnf conf.IServerConf) (bool, error) {
	comparer := conf.NewComparer(w.currentConf, cnf)
	if !comparer.IsChanged() {
		return false, nil
	}
	if comparer.IsValueChanged("status", "address", "engines", "host", "readTimeout", "writeTimeout", "readHeaderTimeout") {
		return true, nil
	}
	if ok, err := comparer.IsRequiredSubConfChanged("router"); err != nil || ok {
		return ok, fmt.Errorf("路由未配置或配置有误:%s(%+v)", cnf.GetServerName(), err)
	}
	if ok := comparer.IsSubConfChanged("header"); ok {
		return ok, nil
	}
	if ok := comparer.IsSubConfChanged("view"); ok {
		return ok, nil
	}
	return false, nil
}

//SetConf 设置配置参数
func (w *WebResponsiveServer) SetConf(restart bool, conf conf.IServerConf) (err error) {
	if err = w.ApiResponsiveServer.SetConf(restart, conf); err != nil {
		return err
	}
	//设置metric
	var ok bool
	if ok, err = SetView(w.webServer, conf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, getEnableName(ok), "view设置")
	return nil
}
