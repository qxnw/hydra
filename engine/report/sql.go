package report

import (
	"fmt"

	"strings"

	"strconv"

	"github.com/qxnw/hydra/context"
)

func (s *reportProxy) sqlQueryHandle(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	sql, err := ctx.Input.GetVarParamByArgsName("sql", "sql")
	if err != nil || sql == "" {
		err = fmt.Errorf("var.sql参数未配置:%v", err)
		return
	}
	influxDB, err := ctx.Influxdb.GetClient()
	if err != nil {
		return
	}

	err = ctx.Input.CheckArgs("measurement", "tags", "fields")
	if err != nil {
		return
	}

	measurement := ctx.Input.GetArgValue("measurement")
	tagNames := strings.Split(ctx.Input.GetArgValue("tags"), ",")
	filedNames := strings.Split(ctx.Input.GetArgValue("fields"), ",")

	data, err := ctx.DB.GetDataRows([]string{sql}, map[string]interface{}{})
	if err != nil {
		err = fmt.Errorf("数据查询出错:sql:%v,err:%v", sql, err)
		return
	}
	for _, row := range data {
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		for _, v := range tagNames {
			if !row.Has(v) {
				err = fmt.Errorf("返回的数据集中未包含%s字段", v)
				return
			}
			tags[v] = row.GetString(v)
		}
		for _, v := range filedNames {
			if !row.Has(v) {
				err = fmt.Errorf("返回的数据集中未包含%s字段", v)
				return
			}
			f, err := strconv.ParseFloat(row.GetString(v), 64)
			if err != nil {
				err = fmt.Errorf("字段%s不是有效的float类型段", v)
				return response, err
			}
			fields[v] = f
		}
		err = influxDB.Send(measurement, tags, fields)
		if err != nil {
			return
		}
	}
	return
}
