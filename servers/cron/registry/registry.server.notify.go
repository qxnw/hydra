package registry

import (
	"fmt"
	"strings"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//Notify 服务器配置变更通知
func (w *RegistryServer) Notify(cnf xconf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	sconf := conf.NewRegistryConf(w.conf, cnf)
	if w.conf.GetVersion() == cnf.GetVersion() {
		w.server.Infof("%s:配置未变化", sconf.GetFullName())
		return nil
	}
	//检查是否需要重启服务器
	restart, err := w.NeedRestart(sconf.GetFullName(), cnf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		err = w.Restart(cnf)
		if err != nil {
			return err
		}
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	err = w.SetConf(sconf)
	if err != nil {
		return err
	}
	if restart {
		w.Logger.Infof("%s:启动成功", sconf.GetFullName())
	}
	return nil

}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *RegistryServer) NeedRestart(name string, conf xconf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	if conf.String("sharding") != w.conf.String("sharding") {
		return true, nil
	}

	redis, err := conf.GetNodeWithSectionName("redis", "#@path/redis")
	if err != nil {
		redis = xconf.NewJSONConfWithEmpty()
	}
	r, err := w.conf.GetNodeWithSectionName("redis", "#@path/redis")
	if err != nil {
		r = xconf.NewJSONConfWithEmpty()
	}
	//检查redis是否变化，已变化则需要重启服务
	if r.GetVersion() != redis.GetVersion() {
		return true, nil
	}

	tasks, err := conf.GetNodeWithSectionName("task", "#@path/task")
	if err != nil {
		return false, fmt.Errorf("task未配置或配置有误:%s(%+v)", name, err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("task", "#@path/task"); err != nil || r.GetVersion() != tasks.GetVersion() {
		return true, nil
	}
	return false, nil

}
