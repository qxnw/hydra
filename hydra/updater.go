package hydra

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

func update(domain string, address string, log *logger.Logger) (err error) {
	rgst, err := registry.NewRegistryWithAddress(address, log)
	if err != nil {
		err = fmt.Errorf("初始化注册中心失败：%s:%v", address, err)
		return err
	}
	path := fmt.Sprintf("%s/var/global/logger", domain)
	buff, err := r.getConfig(rgst, path)
	if err != nil {
		return err
	}
	loggerConf, err := conf.NewJSONConfWithJson(string(buff), 0, nil)
	if err != nil {
		err = fmt.Errorf("rpc日志配置错误:%s,%v", string(buff), err)
		return
	}
	return
}
