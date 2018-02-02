package registry

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
)

//Notify 服务器配置变更通知
func (w *RegistryServer) Notify(conf conf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf.GetVersion() == conf.GetVersion() {
		w.server.Infof("%s.%s:配置未变化", conf.String("name"), conf.String("type"))
		return nil
	}
	//检查是否需要重启服务器
	restart, err := w.NeedRestart(conf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		return w.Restart(conf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	return w.SetConf(NewRegistryConf(w.conf, conf))
}

//NeedRestart 检查配置判断是否需要重启服务器
func (w *RegistryServer) NeedRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	if w.conf.String("address") != conf.String("address") {
		return true, nil
	}
	if w.conf.String("host") != conf.String("host") {
		return true, nil
	}

	routers, err := conf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return false, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("router", "#@path/router"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	headers, err := conf.GetNodeWithSectionName("header", "#@path/header")
	if err != nil {
		return false, nil
	}
	//检查头配置是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("header", "#@path/header"); err != nil || r.GetVersion() != headers.GetVersion() {
		return true, nil
	}
	return false, nil

}
