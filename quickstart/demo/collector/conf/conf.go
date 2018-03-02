package conf

import (
	"github.com/qxnw/hydra/component"
)

//Config 项目基础配置
type Config struct {
	RegistryAddress string ``
	AESKey          string `json:"aesKey"`
	SmsSendLimit    int
	SmsErrorLimit   int
	SmsSendStatus   int
	PrepareLimit    int    `json:"prepareLimit"`
	DbName          string `json:"db"`
	CacheName       string `json:"cache"`
	MQName          string `json:"mq"`
	SSOMainURL      string `json:"sso"`
}

var Conf *Config

func Init(c component.IContainer) error {
	cnf := &Config{}
	sc := component.NewStandardConf(c)
	n, err := sc.GetConf(cnf)
	if err != nil {
		return err
	}
	Conf = n.(*Config)
	return nil
}
