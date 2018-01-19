package component

import (
	"fmt"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

//IComponentDB Component DB
type IComponentDB interface{
	GetDefaultDB() (c *db.DB, err error)
	GetDB(name string) (d *db.DB, err error)
}
//StandardDB db
type StandardDB struct {
	engine.IContainer
	name string
}

//NewStandardDB 创建DB
func NewStandardDB(c engine.IContainer, name ...string) *StandardDB {
	if len(name) > 0 {
		return &StandardDB{IContainer: c, name: name[0]}
	}
	return &StandardDB{IContainer: c, name: "db"}
}

//GetDefaultDB 获取默然配置DB
func (s *StandardDB) GetDefaultDB() (c *db.DB, err error) {
	return s.GetDB(s.name)
}

//GetDB 获取数据库操作对象
func (s *StandardDB) GetDB(name string) (d *db.DB, err error) {
	_, dbc, err := dbMap.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		content, err := s.IContainer.GetVarParam("db", name)
		if err != nil {
			return nil, err
		}
		configMap, err := jsons.Unmarshal([]byte(content))
		if err != nil {
			return nil, err
		}
		provider, ok := configMap["provider"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含provider节点:var/db/%s", name)
		}
		connString, ok := configMap["connString"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含connString节点:var/db/%s", name)
		}
		p, c, m := provider.(string), connString.(string), types.ToInt(configMap["max"], 2)
		d, err = db.NewDB(p, c, m)
		if err != nil {
			err = fmt.Errorf("创建DB失败:%s,%s,%d,err:%v", p, c, m, err)
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

var dbMap cmap.ConcurrentMap

func init() {
	dbMap = cmap.New(2)
}
