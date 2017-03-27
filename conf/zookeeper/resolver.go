package zookeeper

import (
	"errors"
	"strings"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/zk"
)

//ZookeeperConfResolver zookeeper配置引擎
type ZookeeperConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *ZookeeperConfResolver) Resolve(args ...string) (conf.ConfWatcher, error) {
	if len(args) < 3 {
		return nil, errors.New("输入参数不能为空")
	}
	servers := args[0]
	domain := args[1]
	tag := args[2]
	client, err := zk.New(strings.Split(servers, ";"), time.Second*3)
	if err != nil {
		return nil, errors.New("无法创建zookeeper")
	}
	return NewZookeeperConfWatcher(domain, tag, client), nil
}
func init() {
	conf.Register("zk", &ZookeeperConfResolver{})
}
