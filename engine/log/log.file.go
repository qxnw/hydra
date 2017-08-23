package log

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/logger"
)

func (s *logProxy) logFileInfoHandle(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	if ctx.Input.Body == "" {
		err = fmt.Errorf("未设置日志内容")
		return
	}
	name, err = ctx.Input.GetArgByName("name")
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
func (s *logProxy) logFileErrorHandle(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	if ctx.Input.Body == "" {
		err = fmt.Errorf("未设置日志内容")
		return
	}
	name, err = ctx.Input.GetArgByName("name")
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
