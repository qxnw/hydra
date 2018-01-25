package influx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/types"
)

func getQueryParams(ctx *context.Context) (sql string, err error) {
	body, err := ctx.Request.Ext.GetBody()
	if err != nil {
		return
	}
	sql, err = ctx.Request.Form.Get("q")
	if err != nil && !types.IsEmpty(body) {
		sql = body
		if !strings.HasPrefix(sql, "select") && !strings.HasPrefix(sql, "show") {
			err = fmt.Errorf("输入的SQL语句必须是select或show开头，(%s)", sql)
			return
		}
		return sql, nil
	}
	if err != nil {
		err = errors.New("form中未包含select标签")
		return
	}
	if !strings.HasPrefix(sql, "select") && !strings.HasPrefix(sql, "select") {
		err = fmt.Errorf("输入的SQL语句必须是select或show开头，(%s)", sql)
		return
	}
	return sql, nil
}

//Query 查询influxdb数据
func Query(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		sql, err := getQueryParams(ctx)
		if err != nil {
			return
		}
		client, err := c.GetInflux("influx")
		if err != nil {
			return
		}
		r, err := client.Query(sql)
		if err != nil {
			err = fmt.Errorf("sql执行出错:%s，(err:%v)", sql, err)
			return
		}
		response.Success(r)
		return
	}
}
