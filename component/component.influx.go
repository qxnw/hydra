package component

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/hydra/engine"
)

//IComponentInfluxDB Component DB
type IComponentInfluxDB interface{
	GetDefaultDB() (c *influxdb.InfluxClient, err error)
	GetDB(name string) (d *influxdb.InfluxClient, err error)
}

//StandardInfluxDB db
type StandardInfluxDB struct {
	engine.IContainer
	name string
}

//NewStandardInfluxDB 创建DB
func NewStandardInfluxDB(c engine.IContainer, name ...string) *StandardInfluxDB {
	if len(name) > 0 {
		return &StandardInfluxDB{IContainer: c, name: name[0]}
	}
	return &StandardInfluxDB{IContainer: c, name: "influxdb"}
}

//GetDefaultDB 获取默然配置DB
func (s *StandardInfluxDB) GetDefaultDB() (c *influxdb.InfluxClient, err error) {
	return s.GetDB(s.name)
}


func (s *StandardInfluxDB) GetDB(name ...string) (*influxdb.InfluxClient, error) {
	influxName := "influxdb"
	if len(name) > 0 {
		influxName = name[0]
	}
	content, err := s.IContainer.GetVarParam("influxdb", influxName)
	if err != nil {
		return nil, err
	}
	_, client, err := influxdbCache.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
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
