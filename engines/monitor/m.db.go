package monitor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
)

//CollectDBValue 收集数据库值
func CollectDBValue(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		sql, err := ctx.Request.Ext.GetVarParam("sql", ctx.Request.Setting.GetString("sql"))
		if err != nil || sql == "" {
			return
		}
		argsName, _ := ctx.Request.Setting.Get("sql")
		sqlNames := strings.Split(argsName, "/")
		db, err := c.GetDB("db")
		if err != nil {
			return
		}
		data, _, _, err := db.Scalar(sql, map[string]interface{}{})
		if err != nil {
			err = fmt.Errorf("数据查询出错:sql:%v,err:%v", sql, err)
			return
		}
		if data == nil {
			data = 0
		}
		value, err := strconv.Atoi(fmt.Sprintf("%v", data))
		if err != nil {
			err = fmt.Errorf("sql:%s返回结果不是有效的数字", sql)
			return
		}
		ip := xnet.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateDBStatus(ctx, int64(value), "server", ip, "name", sqlNames[len(sqlNames)-1])
		response.SetError(0, err)
		return
	}
}
