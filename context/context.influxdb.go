package context

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
)

type ContextInfluxdb struct {
	ctx *Context
}

//Reset 重置context
func (c *ContextInfluxdb) Reset(ctx *Context) {
	c.ctx = ctx
}
func (s *ContextInfluxdb) GetClient(name ...string) (*influxdb.InfluxClient, error) {
	influxName := "influxdb"
	if len(name) > 0 {
		influxName = name[0]
	}
	content, err := s.ctx.Input.GetVarParamByArgsName("influxdb", influxName)
	if err != nil {
		return nil, err
	}
	_, client, err := influxdbCache.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		client, err := influxdb.NewInfluxClient(host, dataBase, cnf.String("userName"), cnf.String("password"))
		if err != nil {
			return nil, fmt.Errorf("初始化失败(err:%v)", err)
		}
		return client, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*influxdb.InfluxClient), err

}

var influxdbCache cmap.ConcurrentMap

func init() {
	influxdbCache = cmap.New(2)
}
