package mqc

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
)

//Notify 服务器配置变更通知
func (w *MqcResponsiveServer) Notify(nConf conf.IServerConf) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	//检查是否需要重启服务器
	restart, err := w.NeedRestart(nConf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		servers.Tracef(w.Infof, "%s:重启服务", nConf.GetServerName())
		w.currentConf = nConf
		return w.Restart(nConf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	if err = w.SetConf(false, nConf); err != nil {
		return err
	}
	w.engine.UpdateVarConf(nConf)
	w.currentConf = nConf
	return nil
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *MqcResponsiveServer) NeedRestart(cnf conf.IServerConf) (bool, error) {
	comparer := conf.NewComparer(w.currentConf, cnf)
	if !comparer.IsChanged() {
		return false, nil
	}
	if comparer.IsValueChanged("status", "engines", "sharding") {
		return true, nil
	}
	if ok, err := comparer.IsRequiredSubConfChanged("server"); err != nil || ok {
		return ok, fmt.Errorf("server未配置或配置有误:%s(%+v)", cnf.GetServerName(), err)
	}
	if ok, err := comparer.IsRequiredSubConfChanged("queue"); err != nil || ok {
		return ok, fmt.Errorf("queue未配置或配置有误:%s(%+v)", cnf.GetServerName(), err)
	}
	return false, nil

}

//SetConf 设置配置参数
func (w *MqcResponsiveServer) SetConf(restart bool, conf conf.IServerConf) (err error) {

	//设置分片数量
	w.shardingCount = conf.GetInt("sharding", 0)

	var ok bool
	//设置task
	if restart {
		if _, err := SetQueues(w.engine, w.server, conf, nil); err != nil {
			return err
		}
	}
	if err != nil {
		return
	}
	//设置metric
	if ok, err = SetMetric(w.server, conf); err != nil {
		return err
	}
	servers.TraceIf(ok, w.Infof, w.Warnf, conf.GetServerName(), getEnableName(ok), "metric设置")

	return nil
}
func getEnableName(b bool) string {
	if b {
		return "启用"
	}
	return "未启用"
}
