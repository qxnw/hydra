package influx

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
)

func (s *influxProxy) getQueryParams(ctx *context.Context) (sql string, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	if ctx.Input.Body != nil {
		sql = ctx.Input.Body.(string)
		/*if !strings.HasPrefix(sql, "SELECT") && !strings.HasPrefix(sql, "show") {
			err = fmt.Errorf("engine:influx.输入的SQL语句必须是select或show开头，(%s)", sql)
			return
		}*/
		return sql, nil
	}
	input := ctx.Input.Input.(transform.ITransformGetter)
	sql, err = input.Get("q")
	if err != nil {
		err = errors.New("engine:influx.form中未包含select标签")
		return
	}
	/*if !strings.HasPrefix(sql, "SELECT") && !strings.HasPrefix(sql, "show") {
		err = fmt.Errorf("engine:influx.输入的SQL语句必须是select或show开头，(%s)", sql)
		return
	}*/
	return sql, nil
}

func (s *influxProxy) query(ctx *context.Context) (r string, err error) {
	sql, err := s.getQueryParams(ctx)
	if err != nil {
		return "", err
	}
	client, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	
	r, err = client.Query(sql)
	if err != nil {
		err = fmt.Errorf("engine:influx.sql执行出错:%s，(err:%v)", sql, err)
		return
	}
	r = "SUCCESS"
	return
}
