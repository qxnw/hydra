package collect

import (
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

var dbCache cmap.ConcurrentMap
var influxdbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New(2)
	influxdbCache = cmap.New(2)
}

func (s *collectProxy) getDB(ctx *context.Context) (*db.DB, error) {
	db, err := ctx.GetArgByName("db")
	if err != nil {
		return nil, err
	}
	content, err := ctx.GetVarParam("db", db)
	if err != nil {
		return nil, fmt.Errorf("无法获取args参数db的值:%s(err:%v)", db, err)
	}
	return getDBFromCache(content)
}
func getDBFromCache(conf string) (*db.DB, error) {
	_, v, err := dbCache.SetIfAbsentCb(conf, func(input ...interface{}) (interface{}, error) {
		config := input[0].(string)
		configMap, err := jsons.Unmarshal([]byte(conf))
		if err != nil {
			return nil, err
		}
		provider, ok := configMap["provider"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含provider节点:%s", config)
		}
		connString, ok := configMap["connString"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含connString节点:%s", config)
		}
		return db.NewDB(provider.(string), connString.(string), types.ToInt(configMap["max"], 2))
	}, conf)
	if err != nil {
		return nil, err
	}
	return v.(*db.DB), nil
}
