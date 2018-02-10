package cron

import (
	"fmt"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/responsive"
)

//Notify 服务器配置变更通知
func (w *CronResponsiveServer) Notify(conf xconf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	nConf := w.currentConf.CopyNew(conf)
	if !nConf.IsChanged() {
		w.Infof("%s:配置未变化", nConf.GetFullName())
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
	return w.SetConf(nConf)
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *CronResponsiveServer) NeedRestart(conf *responsive.ResponsiveConf) (bool, error) {
	if conf.IsValueChanged("status", "engines", "sharding") {
		return true, nil
	}
	if ok, err := conf.IsRequiredNodeChanged("task"); err != nil || ok {
		return ok, fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.GetFullName(), err)
	}
	if ok := conf.IsNodeChanged("redis"); ok {
		return true, nil
	}
	return false, nil

}

//SetConf 设置配置参数
func (w *CronResponsiveServer) SetConf(conf *responsive.ResponsiveConf) (err error) {
	//检查版本号
	if !conf.IsChanged() {
		return nil
	}
	//检查服务器状态
	if conf.IsStoped() {
		return fmt.Errorf("%s:配置为:stop", conf.GetFullName())
	}

	//设置分片数量
	w.shardingCount = conf.GetInt("sharding", 0)

	var ok bool
	//设置task
	if ok, err = conf.IsRequiredNodeChanged("task"); err == nil && ok {
		if _, err := conf.SetTasks(w.engine, w.server, nil); err != nil {
			err = fmt.Errorf("%s:路由配置有误:%v", conf.GetFullName(), err)
			return err
		}
	}

	//设置metric
	if ok, err = conf.SetMetric(w.server); err != nil {
		err = fmt.Errorf("%s:metric配置有误:%v", conf.GetFullName(), err)
		return err
	}
	w.Infof("%s:%smetric设置", conf.GetFullName(), getEnableName(ok))

	return nil
}
func getEnableName(b bool) string {
	if b {
		return "启用"
	}
	return "禁用"
}
