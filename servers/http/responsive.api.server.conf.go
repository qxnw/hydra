package http

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
)

//Notify 服务器配置变更通知
func (w *ApiResponsiveServer) Notify(conf conf.IServerConf) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	//检查是否需要重启服务器
	restart, err := w.NeedRestart(conf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		servers.Tracef(w.Infof, "%s:重启服务", conf.GetServerName())
		w.currentConf = conf
		return w.Restart(conf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	if err = w.SetConf(false, conf); err != nil {
		return err
	}
	w.engine.UpdateVarConf(conf)
	w.currentConf = conf
	return nil
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *ApiResponsiveServer) NeedRestart(cnf conf.IServerConf) (bool, error) {
	comparer := conf.NewComparer(w.currentConf, cnf)
	if !comparer.IsChanged() {
		return false, nil
	}
	if comparer.IsValueChanged("status", "address", "engines", "host", "readTimeout", "writeTimeout", "readHeaderTimeout") {
		return true, nil
	}
	ok, err := comparer.IsRequiredSubConfChanged("router")
	if ok {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("路由未配置或配置有误:%s(%+v)", cnf.GetServerName(), err)
	}
	if ok := comparer.IsSubConfChanged("header"); ok {
		return ok, nil
	}
	if ok := comparer.IsSubConfChanged("circuit"); ok {
		return ok, nil
	}
	return false, nil
}

//SetConf 设置配置参数
func (w *ApiResponsiveServer) SetConf(restart bool, cnf conf.IServerConf) (err error) {
	if !cnf.HasSubConf("router", "static") {
		err = errors.New("路由或静态文件未配置")
		return err
	}

	var ok bool
	//设置路由
	if restart {
		if _, err := SetHttpRouters(w.engine, w.server, nil); err != nil {
			return err
		}
	}
	//设置静态文件
	if ok, err = SetStatic(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "静态文件")

	//设置请求头
	if ok, err = SetHeaders(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "header设置")

	//设置熔断配置
	if ok, err = SetCircuitBreaker(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "熔断设置")

	//设置jwt安全认证
	if ok, err = SetJWT(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "jwt设置")

	//设置ajax请求
	if ok, err = SetAjaxRequest(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "ajax请求限制设置")

	//设置metric
	if ok, err = SetMetric(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "metric设置")

	//设置host
	if ok, err = SetHosts(w.server, cnf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, cnf.GetServerName(), getEnableName(ok), "host设置")

	return nil
}
func getEnableName(b bool) string {
	if b {
		return "启用"
	}
	return "未启用"
}
