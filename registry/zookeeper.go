package registry

import (
	"time"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/zk"
)

type zkRegistryResolver struct {
}

func (z *zkRegistryResolver) Resolve(servers []string, log *logger.Logger) (Registry, error) {
	zclient, err := zk.NewWithLogger(servers, time.Second, log)
	if err != nil {
		return nil, err
	}
	err = zclient.Connect()
	return zclient, err
}

func init() {
	Register("zk", &zkRegistryResolver{})
}
