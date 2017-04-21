package cluster

import (
	"time"

	"fmt"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/zk"
)

func getRegistry(name string, log *logger.Logger, servers []string) (r registry.Registry, err error) {
	switch name {
	case "zk":
		zclient, err := zk.NewWithLogger(servers, time.Second, log)
		if err != nil {
			return nil, err
		}
		err = zclient.Connect()
		r = zclient
		return r, err
	}
	return nil, fmt.Errorf("不支持的注册中心:%s", name)

}
