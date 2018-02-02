package registry

import (
	"strings"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//Notify 服务器配置变更通知
func (w *RegistryServer) Notify(cnf xconf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	//检查是否需要重启服务器
	restart, err := w.NeedRestart(cnf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		return w.Restart(cnf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	return w.SetConf(conf.NewRegistryConf(w.Conf, cnf))
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *RegistryServer) NeedRestart(cnf xconf.Conf) (bool, error) {
	if w.Conf.GetVersion() == cnf.GetVersion() {
		w.Infof("%s.%s:配置未变化", cnf.String("name"), cnf.String("type"))
		return false, nil
	}

	if !strings.EqualFold(cnf.String("status"), w.Conf.String("status")) {
		return true, nil
	}
	if w.Conf.String("address") != cnf.String("address") {
		return true, nil
	}
	if w.Conf.String("host") != cnf.String("host") {
		return true, nil
	}

	headers, err := cnf.GetNodeWithSectionName("header", "#@path/header")
	if err != nil {
		return false, nil
	}
	//检查头配置是否变化，已变化则需要重启服务
	if r, err := w.Conf.GetNodeWithSectionName("header", "#@path/header"); err != nil || r.GetVersion() != headers.GetVersion() {
		return true, nil
	}
	return false, nil

}
