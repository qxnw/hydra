package report

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
)

var dbCache cmap.ConcurrentMap
var influxdbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New(2)
	influxdbCache = cmap.New(2)
}

func (s *collectProxy) getDB(ctx *context.Context) (*db.DB, error) {
	argsMap := ctx.GetArgs()
	db, ok := argsMap["db"]
	if db == "" || !ok {
		return nil, fmt.Errorf("args配置错误，缺少db参数:%v", ctx.GetArgs())
	}
	content, err := s.getVarParam(ctx, "db", db)
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
		return db.NewDB(provider.(string), connString.(string), 10)
	}, conf)
	if err != nil {
		return nil, err
	}
	return v.(*db.DB), nil
}
func (s *collectProxy) getVarParam(ctx *context.Context, tpName string, name string) (string, error) {
	func_var := ctx.GetExt()["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		return f(tpName, name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
