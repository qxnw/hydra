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
	data, err := s.getVarParam(ctx, "setting", ctx.GetArgs()["setting"])
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
	if _, ok := ctx.GetArgs()["url"]; !ok {
		err = fmt.Errorf("args中必须包含参数url,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.httpCollect(ctx, []interface{}{ctx.GetArgs()["url"]}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) tcpHandle(ctx *context.Context) (r string, st int, err error) {
	if _, ok := ctx.GetArgs()["host"]; !ok {
		err = fmt.Errorf("args中必须包含参数host,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.tcpCollect(ctx, []interface{}{ctx.GetArgs()["host"]}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) registryHandle(ctx *context.Context) (r string, st int, err error) {
	if _, ok := ctx.GetArgs()["path"]; !ok {
		err = fmt.Errorf("args中必须包含参数path,%v", ctx.GetArgs())
		return
	}
	if _, ok := ctx.GetArgs()["min"]; !ok {
		err = fmt.Errorf("args中必须包含参数min,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.registryCollect(ctx, []interface{}{ctx.GetArgs()["path"], ctx.GetArgs()["min"]}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) cpuHandle(ctx *context.Context) (r string, st int, err error) {
	if _, ok := ctx.GetArgs()["max"]; !ok {
		err = fmt.Errorf("args中必须包含参数max,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.cpuCollect(ctx, []interface{}{ctx.GetArgs()["max"]}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) memHandle(ctx *context.Context) (r string, st int, err error) {
	if _, ok := ctx.GetArgs()["max"]; !ok {
		err = fmt.Errorf("args中必须包含参数max,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.memCollect(ctx, []interface{}{ctx.GetArgs()["max"]}, influxdb)
	if r == "NONEED" {
		st = 204
	}
	return
}
func (s *collectProxy) diskHandle(ctx *context.Context) (r string, st int, err error) {
	if _, ok := ctx.GetArgs()["max"]; !ok {
		err = fmt.Errorf("args中必须包含参数max,%v", ctx.GetArgs())
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err = s.diskCollect(ctx, []interface{}{ctx.GetArgs()["max"]}, influxdb)
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
	db, ok := ctx.GetArgs()["influxdb"]
	if db == "" || !ok {
		return nil, fmt.Errorf("args配置错误，缺少influxdb参数:%v", ctx.GetArgs())
	}
	content, err := s.getVarParam(ctx, "db", db)
	if err != nil {
		return nil, fmt.Errorf("无法获取args参数influxdb的值:%s(err:%v)", db, err)
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
