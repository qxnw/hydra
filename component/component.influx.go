package component

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
)

//IComponentInfluxDB Component DB
type IComponentInfluxDB interface {
	GetDefaultInflux() (c *influxdb.InfluxClient, err error)
	GetInflux(name string) (d *influxdb.InfluxClient, err error)
	Close() error
}

//StandardInfluxDB db
type StandardInfluxDB struct {
	IContainer
	name          string
	influxdbCache cmap.ConcurrentMap
}

//NewStandardInfluxDB 创建DB
func NewStandardInfluxDB(c IContainer, name ...string) *StandardInfluxDB {
	if len(name) > 0 {
		return &StandardInfluxDB{IContainer: c, name: name[0], influxdbCache: cmap.New(2)}
	}
	return &StandardInfluxDB{IContainer: c, name: "influxdb", influxdbCache: cmap.New(2)}
}

//GetDefaultInflux 获取默认influxdb
func (s *StandardInfluxDB) GetDefaultInflux() (c *influxdb.InfluxClient, err error) {
	return s.GetInflux(s.name)
}

//GetInflux get influxdb
func (s *StandardInfluxDB) GetInflux(name string) (*influxdb.InfluxClient, error) {
	content, err := s.IContainer.GetVarParam("influxdb", name)
	if err != nil {
		return nil, err
	}
	_, client, err := s.influxdbCache.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
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

//Close 释放所有缓存配置
func (s *StandardInfluxDB) Close() error {
	s.influxdbCache.RemoveIterCb(func(k string, v interface{}) bool {
		v.(*influxdb.InfluxClient).Close()
		return true
	})
	return nil
}
