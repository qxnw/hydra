package mock

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
)

func (s *mockProxy) rawMockHandle(ctx *context.Context) (r string, t int, headerMap map[string]interface{}, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}

	args, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("args类型错误必须为map[string]string:%v", ctx.Input)
		return
	}
	setting, ok := args["setting"]
	if !ok {
		err = fmt.Errorf("args配置错误，未指定setting参数的值:%v", args)
		return
	}

	content, err := s.getVarParam(ctx, "setting", setting)
	if err != nil {
		err = fmt.Errorf("args配置错误，args.setting配置的节点:%s获取失败(err:%v)", setting, err)
		return
	}
	paraTransform := transform.NewGetter(ctx.Input.Params.(transform.ITransformGetter))
	paraTransform.Append(ctx.Input.Input.(transform.ITransformGetter))
	r = paraTransform.Translate(content)
	headerMap = make(map[string]interface{})
	headerMap["Content-Type"] = "text/plain"
	header, ok := args["header"]
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
	funcVar := ctx.Ext["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f(tp, name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
