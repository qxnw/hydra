package context

import (
	"errors"
	"fmt"

	"github.com/qxnw/lib4go/utility"
)

type extParams struct {
	ext map[string]interface{}
}

func (w *extParams) Get(name string) (interface{}, bool) {
	v, ok := w.ext[name]
	return v, ok
}
func (w *extParams) GetBodyMap(encoding ...string) map[string]string {
	content, err := w.GetBody(encoding...)
	if err != nil {
		return make(map[string]string)
	}
	mSetting, err := utility.GetMapWithQuery(content)
	if err != nil {
		return make(map[string]string)
	}
	return mSetting

}
func (w *extParams) GetBody(encoding ...string) (string, error) {
	e := "utf-8"
	if len(encoding) > 0 {
		e = encoding[0]
	}
	if fun, ok := w.ext["__func_body_get_"].(func(ch string) (string, error)); ok {
		return fun(e)
	}
	return "", fmt.Errorf("无法根据%s格式转换数据", e)
}

func (w *extParams) getVarParam() (func(c string, n string) (string, error), error) {
	funcVar := w.ext["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_var_get_传入类型错误")
}

//GetVarParam 获取var参数值，需提供在ext中提供__func_var_get_
func (w *extParams) GetVarParam(tpName string, name string) (string, error) {
	f, err := w.getVarParam()
	if err != nil {
		return "", err
	}
	return f(tpName, name)
}

//GetJWTBody 获取jwt存储数据
func (w *extParams) GetJWTBody() interface{} {
	return w.ext["__jwt_"]
}

//GetUUID
func (w *extParams) GetUUID() string {
	return w.ext["hydra_sid"].(string)
}
