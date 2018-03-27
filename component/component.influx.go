package component

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/influxdb"
)

//IComponentInfluxDB Component DB
type IComponentInfluxDB interface {
	GetDefaultInflux() (c influxdb.IInfluxClient, err error)
	GetInflux(name string) (d influxdb.IInfluxClient, err error)
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
func (s *StandardInfluxDB) GetDefaultInflux() (c influxdb.IInfluxClient, err error) {
	return s.GetInflux(s.name)
}

//GetInflux get influxdb
func (s *StandardInfluxDB) GetInflux(name string) (influxdb.IInfluxClient, error) {
	influxDbConf, err := s.IContainer.GetVarConf("influxdb", name)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("../var/influxdb/%s:%d", name, influxDbConf.GetVersion())
	_, client, err := s.influxdbCache.SetIfAbsentCb(key, func(i ...interface{}) (interface{}, error) {
		cnf := i[0].(*conf.JSONConf)
		var metric *conf.Metric
		if err := cnf.Unmarshal(&metric); err != nil {
			return nil, err
		}
		if b, err := govalidator.ValidateStruct(&metric); !b {
			return nil, err
		}
		return influxdb.NewInfluxClient(metric.Host, metric.DataBase, metric.UserName, metric.Password)

	}, influxDbConf)
	if err != nil {
		err = fmt.Errorf("创建influxdb失败:%s,err:%v", string(influxDbConf.GetRaw()), err)
		return nil, err
	}
	return client.(influxdb.IInfluxClient), err

}

//Close 释放所有缓存配置
func (s *StandardInfluxDB) Close() error {
	s.influxdbCache.RemoveIterCb(func(k string, v interface{}) bool {
		v.(*influxdb.InfluxClient).Close()
		return true
	})
	return nil
}
