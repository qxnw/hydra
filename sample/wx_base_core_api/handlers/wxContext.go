package handlers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/qxnw/hydra_plugin/plugins"
	"github.com/qxnw/lib4go/transform"
)

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &wxContext{}
		},
	}
}

type wxContext struct {
	ctx          plugins.Context
	Input        transform.ITransformGetter
	Params       transform.ITransformGetter
	Body         string
	Args         map[string]string
	func_var_get func(c string, n string) (string, error)
	RPC          plugins.RPCInvoker
}

func (w *wxContext) CheckMustFields(names ...string) error {
	for _, v := range names {
		if _, err := w.Input.Get(v); err != nil {
			err := fmt.Errorf("wx_base_core:输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

func getWXContext(ctx plugins.Context, invoker plugins.RPCInvoker) (wx *wxContext, err error) {
	wx = contextPool.Get().(*wxContext)
	wx.ctx = ctx
	if invoker == nil {
		err = fmt.Errorf("wx_base_core:输入参数rpc.invoker为空")
		wx.Close()
		return
	}
	wx.Input, err = wx.getGetParams(ctx.GetInput())
	if err != nil {
		wx.Close()
		return
	}
	wx.Params, err = wx.getGetParams(ctx.GetParams())
	if err != nil {
		wx.Close()
		return
	}
	wx.Body, err = wx.getGetBody(ctx.GetBody())
	if err != nil {
		wx.Close()
		return
	}
	wx.Args, err = wx.GetArgs(ctx.GetArgs())
	if err != nil {
		wx.Close()
		return
	}
	wx.func_var_get, err = wx.getVarParam(ctx.GetExt())
	if err != nil {
		wx.Close()
		return
	}
	wx.RPC = invoker
	return
}

func (w *wxContext) getVarParam(ext map[string]interface{}) (func(c string, n string) (string, error), error) {
	funcVar := ext["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("wx_base_core:未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("wx_base_core:未找到__func_var_get_传入类型错误")
}
func (w *wxContext) GetArgs(args interface{}) (params map[string]string, err error) {
	params, ok := args.(map[string]string)
	if !ok {
		err = fmt.Errorf("未设置Args参数")
		return
	}
	return
}
func (w *wxContext) getGetBody(body interface{}) (t string, err error) {
	if body == nil {
		return "", errors.New("wx_base_core:body 数据为空")
	}
	t, ok := body.(string)
	if !ok {
		return "", errors.New("wx_base_core:body 不是字符串数据")
	}
	return
}
func (w *wxContext) getGetParams(input interface{}) (t transform.ITransformGetter, err error) {
	if input == nil {
		err = fmt.Errorf("输入参数为空:%v", input)
		return nil, err
	}
	t, ok := input.(transform.ITransformGetter)
	if ok {
		return t, err
	}
	return nil, fmt.Errorf("输入参数为空:%v", input)

}
func (w *wxContext) Close() {
	contextPool.Put(w)
}
