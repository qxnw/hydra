package collect

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) dbCollect(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	title := ctx.Input.GetArgValue("title", "数据库监控服务")
	msg := ctx.Input.GetArgValue("msg", "数据库服务:@host当前值:@current")
	sql, err := ctx.Input.GetVarParamByArgsName("sql", "sql")
	if err != nil || sql == "" {
		return
	}
	max := ctx.Input.GetArgInt("max")
	min := ctx.Input.GetArgInt("min")
	data, err := ctx.DB.Scalar([]string{sql}, map[string]interface{}{})
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
	tf.Set("host", ctx.Input.GetArgValue("db"))
	tf.Set("url", ctx.Input.GetArgValue("sql"))
	tf.Set("value", strconv.Itoa(result))
	tf.Set("current", strconv.Itoa(value))
	tf.Set("level", ctx.Input.GetArgValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("{@host}_{@url}"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "db", tf, result)
	response.SetError(st, err)
	return
}
