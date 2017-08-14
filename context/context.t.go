package context

import (
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

//NewTContext 构建用于测试的context对象
func NewTContext(fn func(f, x string) (string, error)) *Context {
	t := GetContext()
	t.SetInput(transform.New().Data,
		transform.New().Data, "",
		make(map[string]string),
		map[string]interface{}{
			"__func_var_get_": func(f, x string) (string, error) {
				return fn(f, x)
			},
			"__test__":  true,
			"hydra_sid": utility.GetGUID()[0:8],
		})
	return t
}
