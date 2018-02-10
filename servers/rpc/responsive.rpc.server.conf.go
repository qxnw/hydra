package rpc

import (
	"fmt"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
)

//Notify 服务器配置变更通知
func (w *RpcResponsiveServer) Notify(conf xconf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	nConf := w.currentConf.CopyNew(conf)
	if !nConf.IsChanged() {
		servers.Trace(w.Infof, "%s:配置未变化", nConf.GetFullName())
		return nil
	}
	//检查是否需要重启服务器
	restart, err := w.NeedRestart(nConf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		return w.Restart(nConf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	if err = w.SetConf(nConf); err != nil {
		return err
	}
	w.currentConf = nConf
	return nil
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *RpcResponsiveServer) NeedRestart(conf *responsive.ResponsiveConf) (bool, error) {
	if conf.IsValueChanged("status", "address", "engines", "host") {
		return true, nil
	}
	if ok, err := conf.IsRequiredNodeChanged("router"); err != nil || ok {
		return ok, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.GetFullName(), err)
	}
	if ok := conf.IsNodeChanged("header"); ok {
		return ok, nil
	}
	return false, nil

}

//SetConf 设置配置参数
func (w *RpcResponsiveServer) SetConf(conf *responsive.ResponsiveConf) (err error) {
	//检查版本号
	if !conf.IsChanged() {
		return nil
	}
	//检查服务器状态
	if conf.IsStoped() {
		return fmt.Errorf("%s:配置为:stop", conf.GetFullName())
	}
	var ok bool
	//设置路由
	if ok, err = conf.IsRequiredNodeChanged("router"); err == nil && ok {
		if _, err := conf.SetRouters(w.engine, w.server, nil); err != nil {
			err = fmt.Errorf("%s:路由配置有误:%v", conf.GetFullName(), err)
			return err
		}
	}
	if err != nil {
		return
	}
	//设置jwt安全认证
	if ok, err = conf.SetJWT(w.server); err != nil {
		err = fmt.Errorf("%s:header配置有误:%v", conf.GetFullName(), err)
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetFullName(), getEnableName(ok), "jwt设置")
	//设置metric
	if ok, err = conf.SetMetric(w.server); err != nil {
		err = fmt.Errorf("%s:metric配置有误:%v", conf.GetFullName(), err)
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetFullName(), getEnableName(ok), "metric设置")

	//设置host
	if ok, err = conf.SetHosts(w.server); err != nil {
		err = fmt.Errorf("%s:host配置有误:%v", conf.GetFullName(), err)
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetFullName(), getEnableName(ok), "host设置")
	return nil
}
func getEnableName(b bool) string {
	if b {
		return "启用"
	}
	return "未启用"
}
