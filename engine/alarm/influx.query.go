package alarm

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
)

func (s *alarmProxy) getQueryParams(ctx *context.Context) (dbSeting conf.Conf, notifySetting conf.Conf, err error) {
	setting, ok := ctx.GetArgs()["setting"]
	if !ok {
		err = fmt.Errorf("Args参数未配置setting属性")
		return
	}
	content, err := s.getVarParam(ctx, "setting", setting)
	if err != nil {
		err = fmt.Errorf("Args参数的属性setting节点未找到:%v", err)
		return
	}
	form, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}
	dbSeting, err = form.GetSection("db")
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，未配置db节点（err:%v)", content, err)
		return
	}
	notifySetting, err = form.GetSection("notify")
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，未配置db节点（err:%v)", content, err)
		return
	}
	return
}

func (s *alarmProxy) influxQuery(ctx *context.Context, sql string) (rs []map[string]string, err error) {
	client, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	r, err := client.Query(sql)
	if err != nil {
		err = fmt.Errorf("sql执行出错:%s，(err:%v)", sql, err)
		return
	}
	queryResult, err := conf.NewJSONConfWithJson(r, 0, nil, nil)
	if err != nil {
		return
	}
	results, err := queryResult.GetSections("results")
	if err != nil {
		return
	}
	for _, result := range results {
		series, err := result.GetSections("series")
		if err != nil {
			return nil, err
		}
		for _, serie := range series {

			if !serie.Has("columns") || !serie.Has("values") {
				err = fmt.Errorf("influxdb返回数据错误:未包含columns或values列%s，(err:%v)", r, err)
				return nil, err
			}
			columns, err := serie.GetArray("columns")
			if err != nil {
				err = fmt.Errorf("influxdb返回数据错误:columns不是数组%s，(err:%v)", r, err)
				return nil, err
			}
			values, err := serie.GetArray("values")
			if err != nil {
				err = fmt.Errorf("influxdb返回数据错误:values不是数组%s，(err:%v)", r, err)
				return nil, err
			}
			rs = make([]map[string]string, 0, len(values))
			for _, row := range values {
				cols := row.([]interface{})
				rowMap := make(map[string]string)
				for x, c := range cols {
					rowMap[fmt.Sprintf("%v", columns[x])] = fmt.Sprintf("%v", c)
				}
				rs = append(rs, rowMap)
			}
		}
	}

	return
}
