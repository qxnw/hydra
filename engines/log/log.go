package log

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/logger"
)

//WriteInfoLog 写入info类型日志
func WriteInfoLog() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		body, err := ctx.Request.Ext.GetBody()
		if err != nil || body == "" {
			err = fmt.Errorf("未设置日志内容")
			return
		}
		name, err = ctx.Request.Setting.Get("name")
		if err != nil {
			return
		}
		lg := logger.GetSession(name, ctx.Request.Ext.GetUUID())
		lg.Info(body)
		return
	}
}

//WriteErrorLog 写入错误日志
func WriteErrorLog() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		body, err := ctx.Request.Ext.GetBody()
		if err != nil || body == "" {
			err = fmt.Errorf("未设置日志内容")
			return
		}
		name, err = ctx.Request.Setting.Get("name")
		if err != nil {
			return
		}
		lg := logger.GetSession(name, ctx.Request.Ext.GetUUID())
		lg.Error(body)
		return
	}
}
