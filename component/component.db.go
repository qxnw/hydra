package component

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
)

//IComponentDB Component DB
type IComponentDB interface {
	GetDefaultDB() (c *db.DB, err error)
	GetDB(name string) (d *db.DB, err error)
	Close() error
}

//StandardDB db
type StandardDB struct {
	IContainer
	name  string
	dbMap cmap.ConcurrentMap
}

//NewStandardDB 创建DB
func NewStandardDB(c IContainer, name ...string) *StandardDB {
	if len(name) > 0 {
		return &StandardDB{IContainer: c, name: name[0], dbMap: cmap.New(2)}
	}
	return &StandardDB{IContainer: c, name: "db", dbMap: cmap.New(2)}
}

//GetDefaultDB 获取默然配置DB
func (s *StandardDB) GetDefaultDB() (c *db.DB, err error) {
	return s.GetDB(s.name)
}

//GetDB 获取数据库操作对象
func (s *StandardDB) GetDB(name string) (d *db.DB, err error) {
	_, dbc, err := s.dbMap.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		dbJsonConf, err := s.IContainer.GetVarConf("db", name)
		if err != nil {
			return nil, err
		}
		var dbConf conf.DBConf
		if err = dbJsonConf.Unmarshal(&dbConf); err != nil {
			return nil, err
		}
		d, err = db.NewDB(dbConf.Provider, dbConf.ConnString, dbConf.Max)
		if err != nil {
			err = fmt.Errorf("创建DB失败:%s,err:%v", string(dbJsonConf.GetRaw()), err)
			return
		}
		return
	}, name)
	if err != nil {
		return
	}
	d = dbc.(*db.DB)
	return
}

//Close 释放所有缓存配置
func (s *StandardDB) Close() error {
	s.dbMap.RemoveIterCb(func(k string, v interface{}) bool {
		v.(*db.DB).Close()
		return true
	})
	return nil
}
