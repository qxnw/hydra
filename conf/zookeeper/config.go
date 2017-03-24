package zookeeper

import (
	"errors"

	"strings"

	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/zk"
)

//ZookeeperConfAdapter zookeeper配置引擎
type ZookeeperConfAdapter struct {
}

//Parse 从服务器获取数据
func (j *ZookeeperConfAdapter) Parse(args ...string) (c conf.Config, err error) {
	if len(args) < 2 {
		return nil, errors.New("输入参数不能为空")
	}
	servers := args[0]
	path := args[1]
	client, err := zk.New(strings.Split(servers, ";"), time.Second*3)
	if err != nil {
		return
	}
	value, err := client.GetValue(path)
	if err != nil {
		return
	}
	return j.ParseData([]byte(value))
}

//ParseData 转换数据
func (j *ZookeeperConfAdapter) ParseData(data []byte) (conf.Config, error) {
	return conf.NewConfigData("json", data)
}
