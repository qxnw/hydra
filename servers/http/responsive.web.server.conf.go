package http

import (
	"fmt"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
)

//NeedRestart 检查配置判断是否需要重启服务器
func (w *WebResponsiveServer) NeedRestart(conf *responsive.ResponsiveConf) (bool, error) {
	if conf.IsValueChanged("status", "address", "engines", "host", "readTimeout", "writeTimeout", "readHeaderTimeout") {
		return true, nil
	}
	if ok, err := conf.IsRequiredNodeChanged("router"); err != nil || ok {
		return ok, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.GetFullName(), err)
	}
	if ok := conf.IsNodeChanged("header"); ok {
		return ok, nil
	}
	if ok := conf.IsNodeChanged("view"); ok {
		return ok, nil
	}
	return false, nil
}

//SetConf 设置配置参数
func (w *WebResponsiveServer) SetConf(restart bool, conf *responsive.ResponsiveConf) (err error) {
	if err = w.ApiResponsiveServer.SetConf(restart, conf); err != nil {
		return err
	}
	//设置metric
	var ok bool
	if ok, err = conf.SetView(w.webServer); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetFullName(), getEnableName(ok), "view设置")
	return nil
}
