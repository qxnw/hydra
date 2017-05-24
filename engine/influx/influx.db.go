package influx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/influxdb"

	"github.com/qxnw/hydra/conf"
)

func (s *influxProxy) getInfluxClient(ctx *context.Context) (*influxdb.InfluxClient, error) {
	argsMap, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("Args输入参数类型错误不是map[string]string类型:%v", ctx.Input.Args)
	}
	db, ok := argsMap["db"]
	if db == "" || !ok {
		return nil, fmt.Errorf("engine:influxdb.args配置错误，缺少db参数:%v", ctx.Input.Args)
	}
	content, err := s.getVarParam(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("engine:无法获取args参数db的值:%s(err:%v)", db, err)
	}

	_, client, err := s.dbs.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("engine:influxdb.args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("engine:influxdb配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		client, err := influxdb.NewInfluxClient(host, dataBase, cnf.String("userName"), cnf.String("password"))
		if err != nil {
			return nil, fmt.Errorf("engine:influxdb初始化失败(err:%v)", err)
		}
		return client, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*influxdb.InfluxClient), err

}
func (s *influxProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	func_var := ctx.Ext["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("engine:未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		return f("db", name)
	}
	return "", errors.New("engine:未找到__func_var_get_传入类型错误")
}
