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
		if ctx.Input.Body == "" {
			err = fmt.Errorf("未设置日志内容")
			return
		}
		name, err = ctx.Input.GetArgsByName("name")
		if err != nil {
			return
		}
		sessionID, ok := ctx.Input.Ext["hydra_sid"]
		if !ok || sessionID == "" {
			sessionID = logger.CreateSession()
		}
		lg := logger.GetSession(name, sessionID.(string))
		lg.Info(ctx.Input.Body)
		return
	}
}

//WriteErrorLog 写入错误日志
func WriteErrorLog() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		if ctx.Input.Body == "" {
			err = fmt.Errorf("未设置日志内容")
			return
		}
		name, err = ctx.Input.GetArgsByName("name")
		if err != nil {
			return
		}
		sessionID, ok := ctx.Input.Ext["hydra_sid"]
		if !ok || sessionID == "" {
			sessionID = logger.CreateSession()
		}
		lg := logger.GetSession(name, sessionID.(string))
		lg.Error(ctx.Input.Body)
		return
	}
}
