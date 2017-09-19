package registry

import (
	"time"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/zk"
)

//zookeeper 基于zookeeper的注册中心
type zookeeper struct {
}

//Resolve 根据配置生成zookeeper客户端
func (z *zookeeper) Resolve(servers []string, log *logger.Logger) (Registry, error) {
	zclient, err := zk.NewWithLogger(servers, time.Second, log)
	if err != nil {
		return nil, err
	}
	err = zclient.Connect()
	return zclient, err
}

func init() {
	Register("zk", &zookeeper{})
}
