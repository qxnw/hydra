package alarm

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/influxdb/v2"
	"github.com/qxnw/lib4go/jsons"
)

func (s *alarmProxy) getQueryParams(ctx *context.Context) (dbSeting conf.Conf, notifySetting conf.Conf, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	params, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("未设置Args参数")
		return
	}
	setting, ok := params["setting"]
	if !ok {
		err = fmt.Errorf("Args参数未配置setting属性")
		return
	}
	content, err := s.getVarParam(ctx, setting)
	if err != nil {
		err = fmt.Errorf("Args参数的属性setting节点未找到:%v", err)
		return
	}
	form, err := conf.NewJSONConfWithJson(content, 0, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}
	dbSeting, err = form.GetNodeWithSection("db")
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，未配置db节点（err:%v)", content, err)
		return
	}
	notifySetting, err = form.GetNodeWithSection("notify")
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，未配置db节点（err:%v)", content, err)
		return
	}
	return
}

func (s *alarmProxy) influxQuery(ctx *context.Context, sql string) (rs []map[string]string, err error) {
	r, err := client.Query(sql)
	if err != nil {
		err = fmt.Errorf("sql执行出错:%s，(err:%v)", sql, err)
		return
	}
	data, err := jsons.Unmarshal([]byte(r))
	if err != nil {
		err = fmt.Errorf("influxdb返回数据错误:%s，(err:%v)", r, err)
		return
	}
	if data["columns"] == nil || data["values"] == nil {
		err = fmt.Errorf("influxdb返回数据错误:未包含columns或values列%s，(err:%v)", r, err)
		return
	}
	columns, ok := data["columns"].([]interface{})
	if !ok {
		err = fmt.Errorf("influxdb返回数据错误:columns不是数组%s，(err:%v)", r, err)
		return
	}
	values, ok := data["values"].([]interface{})
	if !ok {
		err = fmt.Errorf("influxdb返回数据错误:values不是数组%s，(err:%v)", r, err)
		return
	}
	rs = make([]map[string]string, 0, len(values))
	for i, row := range values {
		cols := row.([]interface{})
		rowMap := make(map[string]string)
		for x, c := range cols {
			rowMap[fmt.Sprintf("%v", columns[x])] = fmt.Sprintf("%v", c)
		}
		rs = append(rs, rowMap)
	}
	return
}
