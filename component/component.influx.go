package component

import (
	"fmt"

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
	influxDbConf, err := s.IContainer.GetVarConf("influxdb", name)
	if err != nil {
		return nil, err
	}
	_, client, err := s.influxdbCache.SetIfAbsentCb(name, func(i ...interface{}) (interface{}, error) {
		cnf := i[0].(*conf.JSONConf)
		var metric *conf.Metric
		if err := cnf.Unmarshal(&metric); err != nil {
			err = fmt.Errorf("../influxdb/%s配置格式有误:%v", name, err)
			return nil, err
		}
		client, err := influxdb.NewInfluxClient(metric.Host, metric.DataBase, metric.UserName, metric.Password)
		if err != nil {
			return nil, fmt.Errorf("初始化失败(err:%v)", err)
		}
		return client, err
	}, influxDbConf)
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
