package registry

import (
	"time"

	"fmt"

	"github.com/qxnw/lib4go/zk"
)

func getRegistry(name string, args ...string) (Registry, error) {
	switch name {
	case "zookeeper":
		return zk.New(args, time.Second)
	}
	return nil, fmt.Errorf("不支持的注册中心:%s", name)

}
