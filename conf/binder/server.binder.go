package binder

import "path/filepath"

type ServerBinder struct {
	API  IBinder
	RPC  IBinder
	WEB  IBinder
	MQC  IBinder
	CRON IBinder
}

func NewServerBinder() *ServerBinder {
	return &ServerBinder{
		API:  NewBinder(),
		RPC:  NewBinder(),
		WEB:  NewBinder(),
		MQC:  NewBinder(),
		CRON: NewBinder(),
	}
}

//GetConfs 获取配置文件
func (s *ServerBinder) GetConfs(t string) map[string]string {
	switch t {
	case "api":
		return s.API.GetNodeConf()
	case "web":
		return s.WEB.GetNodeConf()
	case "rpc":
		return s.RPC.GetNodeConf()
	case "mqc":
		return s.MQC.GetNodeConf()
	case "cron":
		return s.CRON.GetNodeConf()
	}
	return nil
}

func (s *ServerBinder) Bind(platName string, systemName string, serverTypes []string, clusterName string) error {
	for _, tp := range serverTypes {
		mainConf := filepath.Join("/", platName, systemName, tp, clusterName, "conf")
		switch tp {
		case "api":
			if err := s.API.Bind(platName, mainConf); err != nil {
				return err
			}
		case "web":
			if err := s.WEB.Bind(platName, mainConf); err != nil {
				return err
			}
		case "rpc":
			if err := s.RPC.Bind(platName, mainConf); err != nil {
				return err
			}
		case "mqc":
			if err := s.MQC.Bind(platName, mainConf); err != nil {
				return err
			}
		case "cron":
			if err := s.CRON.Bind(platName, mainConf); err != nil {
				return err
			}
		}
	}
	return nil
}
