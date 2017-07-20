package collect

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/influxdb"
)

func (s *collectProxy) getExectionParams(ctx *context.Context) (clct *collector, err error) {
	data, err := ctx.GetVarParamByArgsName("setting", "setting")
	if err != nil {
		err = fmt.Errorf("setting.%s未配置:err:%v", ctx.GetArgs()["setting"], err)
		return
	}

	err = json.Unmarshal([]byte(data), clct)
	if err != nil {
		err = fmt.Errorf("setting.%s的内容必须是json字符串:%v", ctx.GetArgs()["setting"], err)
		return
	}
	return
}
func (s *collectProxy) httpHandle(ctx *context.Context) (r string, st int, err error) {
	url, err := ctx.GetArgByName("url")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.httpCollect(ctx, []interface{}{url}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) tcpHandle(ctx *context.Context) (r string, st int, err error) {
	host, err := ctx.GetArgByName("host")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.tcpCollect(ctx, []interface{}{host}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) registryHandle(ctx *context.Context) (r string, st int, err error) {
	path, err := ctx.GetArgByName("path")
	if err != nil {
		return
	}
	min, err := ctx.GetArgByName("min")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.registryCollect(ctx, []interface{}{path, min}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) cpuHandle(ctx *context.Context) (r string, st int, err error) {
	max, err := ctx.GetArgByName("max")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.cpuCollect(ctx, []interface{}{max}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) memHandle(ctx *context.Context) (r string, st int, err error) {
	max, err := ctx.GetArgByName("max")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.memCollect(ctx, []interface{}{max}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) diskHandle(ctx *context.Context) (r string, st int, err error) {
	max, err := ctx.GetArgByName("max")
	if err != nil {
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.diskCollect(ctx, []interface{}{max}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) dbHandle(ctx *context.Context) (r string, st int, err error) {
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.dbCollect(ctx, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) getInfluxClient(ctx *context.Context) (*influxdb.InfluxClient, error) {
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
