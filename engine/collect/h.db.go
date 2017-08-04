package collect

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) dbCollect(ctx *context.Context) (r string, st int, err error) {

	title := ctx.GetArgValue("title", "数据库监控服务")
	msg := ctx.GetArgValue("msg", "数据库服务:@host,当前值:@current")
	sql, err := ctx.GetVarParamByArgsName("sql", "sql")
	if err != nil || sql == "" {
		return
	}
	smax, ok1 := ctx.GetArgs()["max"]
	smin, ok2 := ctx.GetArgs()["min"]
	if !ok1 && !ok2 {
		err = fmt.Errorf("args未配置max或min")
		return
	}
	max := 0
	min := 0
	if ok1 {
		max, err = strconv.Atoi(smax)
		if err != nil {
			err = fmt.Errorf("args未配置max参数必须是数字:%v", err)
			return
		}
	}
	if ok2 {
		min, err = strconv.Atoi(smin)
		if err != nil {
			err = fmt.Errorf("args未配置min参数必须是数字:%v", err)
			return
		}
	}
	sdb, err := s.getDB(ctx)
	if err != nil {
		err = fmt.Errorf("args数据库db配置有错误:%v", err)
		return
	}
	data, _, _, err := sdb.Scalar(sql, map[string]interface{}{})
	if err != nil {
		err = fmt.Errorf("数据查询出错:sql:%v,err:%v", sql, err)
		return
	}
	if data == nil {
		st = 204
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

	tf.Set("host", ctx.GetArgValue("db"))
	tf.Set("url", ctx.GetArgValue("sql"))
	tf.Set("value", strconv.Itoa(result))
	tf.Set("current", strconv.Itoa(value))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("{@host}_{@url}"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err = s.checkAndSave(ctx, "db", tf, result)
	return
}
