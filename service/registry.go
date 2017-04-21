package service

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/zk"
)

func GetRegistry(name string, args ...string) (registry.Registry, error) {
	switch name {
	case "zk":
		zkclient, err := zk.New(args, time.Second)
		if err != nil {
			return nil, err
		}
		err = zkclient.Connect()
		if err != nil {
			return nil, err
		}
		return zkclient, nil
	}
	return nil, fmt.Errorf("不支持的注册中心:%s", name)
}
