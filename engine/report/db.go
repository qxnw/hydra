package report

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

var dbCache cmap.ConcurrentMap
var influxdbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New(2)
	influxdbCache = cmap.New(2)
}

func (s *reportProxy) getDB(ctx *context.Context) (*db.DB, error) {
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

func (s *reportProxy) getInfluxClient(ctx *context.Context) (*influxdb.InfluxClient, error) {
	content, err := ctx.GetVarParamByArgsName("influxdb", "influxdb")
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
