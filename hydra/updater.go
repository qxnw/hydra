package hydra

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qxnw/hydra/registry"
)

type updaterSetting struct {
	URL     string `json:"url"`
	Version string `json:"v"`
	CRC32   uint32 `json:"crc32"`
}

func (h *Hydra) getPackage(systemName string, version string) (s *updaterSetting, err error) {
	reg, err := registry.NewRegistryWithAddress(h.currentRegistry, h.Logger)
	if err != nil {
		err = fmt.Errorf("注册中心创建失败:%v", err)
		return
	}

	path := fmt.Sprintf("/%s/var/package/%s-%s", h.Domain, systemName, version)
	b, err := reg.Exists(path)
	if err != nil {
		err = fmt.Errorf("无法读取安装包配置:%s", path)
		return
	}
	if !b {
		err = fmt.Errorf("安装包配置不存在:%s", path)
		return
	}
	buff, _, err := reg.GetValue(path)
	if err != nil {
		err = fmt.Errorf("获取更新包配置失败:%v", err)
		return
	}
	s = &updaterSetting{}
	err = json.Unmarshal(buff, s)
	if err != nil {
		err = fmt.Errorf("安装包格式有误:%v", err)
		return
	}
	if s.URL == "" || s.Version == "" || s.CRC32 == 0 {
		err = fmt.Errorf("pack配置有误，参数（url,v,crc32）不能为空")
		return
	}
	return
}

//update 下载安装包并解压到临时目录，停止所有服务器，并拷贝到当前工作目录
func (h *Hydra) updateNow(url string, crc32 uint32) (err error) {
	h.Info("下载安装包:", url)
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("无法下载更新包:%s", url)
		return
	}
	defer resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("无法读取更新包:%v", url)
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("无法读取更新包,状态码:%d", resp.StatusCode)
		return
	}
	if resp.ContentLength == 0 {
		err = fmt.Errorf("无法读取更新包长度:%d", resp.ContentLength)
		return
	}
	updater, err := NewUpdater()
	if err != nil {
		err = fmt.Errorf("无法创建updater:%v", err)
		return
	}
	err = updater.Apply(resp.Body, UpdaterOptions{CRC32: crc32, TargetName: filepath.Base(url)})
	if err != nil {
		if err1 := updater.Rollback(); err1 != nil {
			err = fmt.Errorf("更新失败%+v,回滚失败%v", err, err1)
			return
		}
		err = fmt.Errorf("更新失败,已回滚(err:%v)", err)
		return
	}
	h.Info("更新成功，停止所有服务，并准备重启")
	for _, server := range h.servers {
		server.Shutdown()
	}
	h.healthChecker.Shutdown(time.Second)
	return
}
func (h *Hydra) restartHydra() (err error) {
	args := getExecuteParams(h.inputArgs)
	h.Info("准备重启：", args)
	go func() {
		time.Sleep(time.Second * 5)
		cmd1 := exec.Command("/bin/bash", "-c", args)
		cmd1.Stdout = os.Stdout
		cmd1.Stderr = os.Stderr
		cmd1.Stdin = os.Stdin
		err = cmd1.Start()
		if err != nil {
			return
		}

		h.Info("退出当前程序")
		os.Exit(20)
	}()
	return nil
}
func getExecuteParams(input []string) string {
	args := make([]string, 0, len(input))
	for i, v := range input {
		if i > 0 && strings.HasPrefix(input[i-1], "-") && !strings.HasPrefix(v, "-") {
			args = append(args, fmt.Sprintf(`"%s"`, v))
		} else {
			args = append(args, v)
		}
	}
	return strings.Join(args, " ")
}
