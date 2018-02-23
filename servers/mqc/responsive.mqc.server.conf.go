package mqc

import (
	"errors"
	"fmt"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
)

//Notify 服务器配置变更通知
func (w *MqcResponsiveServer) Notify(conf xconf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	nConf := w.currentConf.CopyNew(conf)
	if !nConf.IsChanged() {
		return nil
	}
	//检查是否需要重启服务器
	restart, err := w.NeedRestart(nConf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		servers.Tracef(w.Infof, "%s:重启服务", nConf.GetFullName())
		w.currentConf = nConf
		return w.Restart(nConf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	if err = w.SetConf(false, nConf); err != nil {
		return err
	}
	w.currentConf = nConf
	return nil
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *MqcResponsiveServer) NeedRestart(conf *responsive.ResponsiveConf) (bool, error) {
	if conf.IsValueChanged("status", "engines", "sharding") {
		return true, nil
	}
	if ok, err := conf.IsRequiredNodeChanged("server"); err != nil || ok {
		return ok, fmt.Errorf("server未配置或配置有误:%s(%+v)", conf.GetFullName(), err)
	}
	if ok, err := conf.IsRequiredNodeChanged("queue"); err != nil || ok {
		return ok, fmt.Errorf("queue未配置或配置有误:%s(%+v)", conf.GetFullName(), err)
	}
	return false, nil

}

//SetConf 设置配置参数
func (w *MqcResponsiveServer) SetConf(restart bool, conf *responsive.ResponsiveConf) (err error) {
	//检查版本号
	if !conf.IsChanged() {
		return nil
	}
	//检查服务器状态
	if conf.IsStoped() {
		return errors.New("配置为:stop")
	}

	//设置分片数量
	w.shardingCount = conf.GetInt("sharding", 0)

	var ok bool
	//设置task
	if ok, err = conf.IsRequiredNodeChanged("queue"); restart || (err == nil && ok) {
		if _, err := conf.SetQueues(w.engine, w.server, nil); err != nil {
			return err
		}
	}
	if err != nil {
		return
	}
	//设置metric
	if ok, err = conf.SetMetric(w.server); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetFullName(), getEnableName(ok), "metric设置")

	return nil
}
func getEnableName(b bool) string {
	if b {
		return "启用"
	}
	return "未启用"
}
