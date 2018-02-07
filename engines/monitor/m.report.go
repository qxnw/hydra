package monitor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//CollectReportValue 收集数据库数值用于报表显示
func CollectReportValue(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		sql, err := c.GetVarParam("sql", ctx.Request.Setting.GetString("sql"))
		if err != nil || sql == "" {
			err = fmt.Errorf("var.sql参数未配置:%v", err)
			return
		}
		influxDB, err := c.GetInflux("influx")
		if err != nil {
			return
		}

		err = ctx.Request.Setting.Check("measurement", "tags", "fields")
		if err != nil {
			return
		}

		measurement := ctx.Request.Setting.GetString("measurement")
		tagNames := strings.Split(ctx.Request.Setting.GetString("tags"), ",")
		filedNames := strings.Split(ctx.Request.Setting.GetString("fields"), ",")

		db, err := c.GetDB("db")
		if err != nil {
			return
		}
		data, _, _, err := db.Query(sql, map[string]interface{}{})
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
}
