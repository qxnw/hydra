package log

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/logger"
)

func (s *logProxy) logFileInfoHandle(ctx *context.Context) (r string, t int, headerMap map[string]interface{}, err error) {
	if b, err := ctx.GetBody(); b == "" || err != nil {
		t = 204
		return "", t, nil, err
	}
	name, ok := ctx.GetArgs()["name"]
	if !ok {
		err = fmt.Errorf("args配置错误，未指定name参数的值:%v", ctx.GetArgs())
		return
	}
	sessionID, ok := ctx.GetExt()["hydra_sid"]
	if !ok || sessionID == "" {
		sessionID = logger.CreateSession()
	}
	lg := logger.GetSession(name, sessionID.(string))
	lg.Info(ctx.GetBody())
	return
}
func (s *logProxy) logFileErrorHandle(ctx *context.Context) (r string, t int, headerMap map[string]interface{}, err error) {
	if b, _ := ctx.GetBody(); b == "" {
		t = 204
		return
	}
	name, ok := ctx.GetArgs()["name"]
	if !ok {
		err = fmt.Errorf("args配置错误，未指定name参数的值:%v", ctx.GetArgs())
		return
	}
	sessionID, ok := ctx.GetExt()["hydra_sid"]
	if !ok || sessionID == "" {
		sessionID = logger.CreateSession()
	}
	lg := logger.GetSession(name, sessionID.(string))
	lg.Error(ctx.GetBody())
	return
}

func (s *logProxy) getVarParam(ctx *context.Context, tp string, name string) (string, error) {
	funcVar := ctx.GetExt()["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f(tp, name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
