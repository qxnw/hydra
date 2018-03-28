package component

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
)

//IComponentDB Component DB
type IComponentDB interface {
	GetDB(names ...string) (d db.IDB, err error)
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

//GetDB 获取数据库操作对象
func (s *StandardDB) GetDB(names ...string) (d db.IDB, err error) {
	name := s.name
	if len(names) > 0 {
		name = names[0]
	}
	dbConf, err := s.IContainer.GetVarConf("db", name)
	if err != nil {
		return nil, fmt.Errorf("../var/db/%s %v", name, err)
	}
	key := fmt.Sprintf("%s:%d", name, dbConf.GetVersion())
	_, dbc, err := s.dbMap.SetIfAbsentCb(key, func(input ...interface{}) (d interface{}, err error) {
		jConf := input[0].(*conf.JSONConf)
		var dbConf conf.DBConf
		if err = jConf.Unmarshal(&dbConf); err != nil {
			return nil, err
		}
		if b, err := govalidator.ValidateStruct(&dbConf); !b {
			return nil, err
		}
		return db.NewDB(dbConf.Provider,
			dbConf.ConnString,
			dbConf.MaxOpen,
			dbConf.MaxIdle,
			dbConf.LefeTime)
	}, dbConf)
	if err != nil {
		err = fmt.Errorf("创建db失败:%s,err:%v", string(dbConf.GetRaw()), err)
		return
	}
	d = dbc.(db.IDB)
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
