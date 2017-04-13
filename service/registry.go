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
		return zk.New(args, time.Second)
	}
	return nil, fmt.Errorf("不支持的注册中心:%s", name)
}
