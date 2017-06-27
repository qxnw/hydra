package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/influxdb"
)

func (s *collectProxy) reportCollect(ctx *context.Context) (r string, st int, err error) {
	data, err := s.getVarParam(ctx, "setting", ctx.GetArgs()["setting"])
	if err != nil {
		err = fmt.Errorf("setting.%s未配置:err:%v", ctx.GetArgs()["setting"], err)
		return
	}
	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	clct := &collector{}
	err = json.Unmarshal([]byte(data), clct)
	if err != nil {
		err = fmt.Errorf("setting.%s的内容必须是json字符串:%v", ctx.GetArgs()["setting"], err)
		return
	}
	resultList := make([]string, 0, 2)
	for _, v := range clct.Collector {
		if collector, ok := s.collector[v.Mode]; ok {
			for _, p := range v.Params {
				result, err := collector(ctx, p, v.Report)
				if err != nil {
					return "", 500, err
				}
				resultList = append(resultList, result...)
			}
		} else {
			err = fmt.Errorf("不支持的模式:%s", v.Mode)
			return
		}
	}
	for _, v := range resultList {
		err := influxdb.SendLineProto(v)
		if err != nil {
			return "", 500, err
		}
	}
	return "SUCCESS", 200, nil

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
