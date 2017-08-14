package context

import (
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

type TContext struct {
	Input      transform.ITransformGetter
	Params     transform.ITransformGetter
	Args       map[string]string
	Body       string
	Ext        map[string]interface{}
	FuncVarGet func(c string, n string) (string, error)
}

func (g *TContext) GetInput() transform.ITransformGetter {
	return g.Input

}
func (g *TContext) GetParams() transform.ITransformGetter {
	return g.Params

}
func (g *TContext) GetArgs() map[string]string {
	return g.Args
}
func (g *TContext) GetBody(encoding ...string) (string, error) {
	return g.Body, nil

}
func (g *TContext) GetExt() map[string]interface{} {
	return g.Ext
}

//NewTContext 构建用于测试的context对象
func NewTContext() *TContext {
	t := &TContext{}
	t.Input = transform.New().Data
	t.Params = transform.New().Data
	t.Ext = map[string]interface{}{
		"__func_var_get_": func(c string, n string) (string, error) {
			if t.FuncVarGet == nil {
				return "", nil
			}
			return t.FuncVarGet(c, n)
		},
		"__test__":  true,
		"hydra_sid": utility.GetGUID()[0:8],
	}
	return t
}
