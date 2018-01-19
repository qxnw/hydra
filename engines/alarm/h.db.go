package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

//DBValueCollect 数据库值收集
func DBValueCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		title := ctx.Input.GetArgsValue("title", "数据库监控服务")
		msg := ctx.Input.GetArgsValue("msg", "数据库服务:@host当前值:@current")
		platform := ctx.Input.GetArgsValue("platform", "----")
		sql, err := ctx.Input.GetVarParamByArgsName("sql", "sql")
		if err != nil || sql == "" {
			return
		}
		max := ctx.Input.GetArgsInt("max")
		min := ctx.Input.GetArgsInt("min")
		db, err := component.NewStandardDB(c, "db").GetDefaultDB()
		if err != nil {
			return
		}
		data, _, _, err := db.Scalar(sql, map[string]interface{}{})
		if err != nil {
			err = fmt.Errorf("数据查询出错:sql:%v,err:%v", sql, err)
			return
		}
		if data == nil {
			response.SetStatus(204)
			return
		}
		value, err := strconv.Atoi(fmt.Sprintf("%v", data))
		if err != nil {
			err = fmt.Errorf("sql:%s返回结果不是有效的数字", sql)
			return
		}
		result := 1 //需要报警
		if ((min > 0 && value >= min) || min == 0) && ((max > 0 && value < max) || max == 0) {
			result = 0 //恢复
		}

		tf := transform.NewMap(map[string]string{})
		tf.Set("host", ctx.Input.GetArgsValue("db"))
		tf.Set("url", ctx.Input.GetArgsValue("sql"))
		tf.Set("value", strconv.Itoa(result))
		tf.Set("current", strconv.Itoa(value))
		tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
		tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("{@host}_{@url}"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, result, "db")
		response.SetError(st, err)
		return
	}
}
