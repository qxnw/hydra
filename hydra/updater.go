package hydra

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/qxnw/hydra/registry"
	"github.com/zkfy/go-update"
)

type updaterSetting struct {
	URL     string `json:"url"`
	Version string `json:"v"`
}

func (h *Hydra) getSetting(version string) (s *updaterSetting, err error) {
	setting := fmt.Sprintf("%s/var/package/%s-%s", h.Domain, h.SystemName, version)
	reg, err := registry.NewRegistry(h.currentRegistry, h.currentRegistryAddress, h.Logger)
	if err != nil {
		err = fmt.Errorf("注册中心创建失败:%v", err)
		return
	}
	buff, _, err := reg.GetValue(setting)
	if err != nil {
		err = fmt.Errorf("获取更新包配置失败:%v", err)
		return
	}
	s = &updaterSetting{}
	err = json.Unmarshal(buff, s)
	if err != nil {
		err = fmt.Errorf("更新包配置有误:%v", err)
		return
	}
	return
}

//update 下载安装包并解压到临时目录，停止所有服务器，并拷贝到当前工作目录
func (h *Hydra) updateNow(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("无法下载更新包:%v", url)
		return
	}
	defer resp.Body.Close()
	var buff []byte
	_, err = resp.Body.Read(buff)
	if err != nil {
		err = fmt.Errorf("无法读取更新包:%v", url)
		return
	}

	err = update.Apply(bytes.NewReader(buff), update.Options{
		Hash:     crypto.SHA256,
		Checksum: buff,
	})
	if err != nil {
		if err1 := update.RollbackError(err); err1 != nil {
			err = fmt.Errorf("更新失败%+v,回滚失败%v", err, err1)
			return
		}
	}
	err = fmt.Errorf("更新失败%+v,已回滚", err)
	return
}
