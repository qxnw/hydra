package mock

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
)

func (s *mockProxy) rawMockHandle(ctx *context.Context) (r string, t int, headerMap map[string]interface{}, err error) {
	setting, ok := ctx.GetArgs()["setting"]
	if !ok {
		err = fmt.Errorf("args配置错误，未指定setting参数的值:%v", ctx.GetArgs())
		return
	}

	content, err := s.getVarParam(ctx, "setting", setting)
	if err != nil {
		err = fmt.Errorf("args配置错误，args.setting配置的节点:%s获取失败(err:%v)", setting, err)
		return
	}
	paraTransform := transform.NewGetter(ctx.GetParams())
	paraTransform.Append(ctx.GetInput())
	r = paraTransform.Translate(content)
	headerMap = make(map[string]interface{})
	headerMap["Content-Type"] = "text/plain"
	header, ok := ctx.GetArgs()["header"]
	if !ok {
		return
	}
	headerContent, err := s.getVarParam(ctx, "header", header)
	if err != nil {
		err = fmt.Errorf("args配置错误，args.header配置的节点:%s获取失败(err:%v)", header, err)
		return
	}

	mapHeader, err := conf.NewJSONConfWithJson(headerContent, 0, nil, nil)
	if err != nil {
		return
	}

	mapHeader.Each(func(k string) {
		headerMap[k] = paraTransform.Translate(mapHeader.String(k))
	})
	return
}
func (s *mockProxy) getVarParam(ctx *context.Context, tp string, name string) (string, error) {
	funcVar := ctx.GetExt()["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f(tp, name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
