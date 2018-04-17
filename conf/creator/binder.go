package creator

import (
	"fmt"
	"path/filepath"
)

type IBinder interface {
	Scan(platName string, systemName string, serverTypes []string, clusterName string) error
	GetMainConf(t string) map[string]string
	GetPlatConf() map[string]string
}
type Binder struct {
	API     IMainBinder
	RPC     IMainBinder
	WEB     IMainBinder
	MQC     IMainBinder
	CRON    IMainBinder
	Plat    IPlatBinder
	binders map[string]IMainBinder
}

func NewBinder() *Binder {
	s := &Binder{}
	s.API = NewMainBinder()
	s.RPC = NewMainBinder()
	s.WEB = NewMainBinder()
	s.MQC = NewMainBinder()
	s.CRON = NewMainBinder()
	s.Plat = NewPlatBinder()
	s.binders = map[string]IMainBinder{
		"api":  s.API,
		"rpc":  s.RPC,
		"web":  s.WEB,
		"mqc":  s.MQC,
		"cron": s.CRON,
	}
	return s
}

//Scan 绑定输入参数
func (s *Binder) Scan(platName string, systemName string, serverTypes []string, clusterName string) error {
	count := s.Plat.NeedScanCount()
	for _, b := range s.binders {
		count += b.NeedScanCount()
	}
	if count == 0 {
		return nil
	}
	var index string
	fmt.Print("当前应用程序启动需要一些关键的参数才能启动，是否立即设置这些参数(yes|NO):")
	fmt.Scan(&index)
	if index != "y" && index != "Y" && index != "yes" && index != "YES" {
		return nil
	}
	for _, v := range serverTypes {
		if err := s.binders[v].Scan(platName, filepath.Join("/", platName, systemName, v, "conf")); err != nil {
			return err
		}
	}
	return s.Plat.Scan(platName)
}

//GetMainConf 获取主配置信息
func (s *Binder) GetMainConf(t string) map[string]string {
	if v, ok := s.binders[t]; ok {
		return v.GetNodeConf()
	}
	return nil
}

//GetPlatConf 获取平台配置信息
func (s *Binder) GetPlatConf() map[string]string {
	return s.Plat.GetNodeConf()
}
