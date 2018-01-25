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
		if err = ctx.Request.Setting.Check("sql"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "数据库监控服务")
		msg := ctx.Request.Setting.GetString("msg", "数据库服务:@host当前值:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		sql, err := ctx.Request.Ext.GetVarParam("sql", ctx.Request.Setting.GetString("sql"))
		if err != nil || sql == "" {
			response.SetStatus(500)
			return
		}
		max := ctx.Request.Setting.GetInt("max")
		min := ctx.Request.Setting.GetInt("min")
		db, err := c.GetDB(ctx.Request.Setting.GetString("db", "db"))
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
		tf.Set("host", ctx.Request.Setting.GetString("db"))
		tf.Set("url", ctx.Request.Setting.GetString("sql"))
		tf.Set("value", strconv.Itoa(result))
		tf.Set("current", strconv.Itoa(value))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
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
