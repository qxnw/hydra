package influx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/types"
)

func (s *influxProxy) getQueryParams(ctx *context.Context) (sql string, err error) {
	body := ctx.Input.Body
	sql, err = ctx.Input.Get("q")
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

func (s *influxProxy) query(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response =context.GetStandardResponse()
	sql, err := s.getQueryParams(ctx)
	if err != nil {
		return
	}
	client, err := ctx.Influxdb.GetClient("influxdb")
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
