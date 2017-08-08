package report

import (
	"fmt"

	"strings"

	"strconv"

	"github.com/qxnw/hydra/context"
)

func (s *reportProxy) sqlQueryHandle(ctx *context.Context) (r string, st int, err error) {
	sql, err := ctx.GetVarParamByArgsName("sql", "sql")
	if err != nil || sql == "" {
		err = fmt.Errorf("var.sql参数未配置:%v", err)
		return
	}
	influxDB, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}

	err = ctx.CheckArgs("measurement", "tags", "fields")
	if err != nil {
		return
	}

	measurement := ctx.GetArgValue("measurement")
	tagNames := strings.Split(ctx.GetArgValue("tags"), ",")
	filedNames := strings.Split(ctx.GetArgValue("fields"), ",")
	sdb, err := s.getDB(ctx)
	if err != nil {
		err = fmt.Errorf("args数据库db配置有错误:%v", err)
		return
	}
	data, _, _, err := sdb.Query(sql, map[string]interface{}{})
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
				return "", 500, err
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
