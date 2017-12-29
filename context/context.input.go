package context

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

type IContextInput interface {
	CheckArgs(names ...string) error
	CheckInput(names ...string) error
	GetVarParam(tpName string, name string) (string, error)
}

//Input 输入参数
type Input struct {
	Input  transform.ITransformGetter
	Params transform.ITransformGetter
	Args   map[string]string
	Ext    map[string]interface{}
	Body   string
}

func (w *Input) Check(checker map[string][]string) (int, error) {
	if err := w.CheckInput(checker["input"]...); err != nil {
		return ERR_NOT_ACCEPTABLE, err
	}
	if err := w.CheckArgs(checker["args"]...); err != nil {
		return ERR_NOT_EXTENDED, err
	}
	return 0, nil
}

//CheckAPISign 指定secret, 所有参数(除sign)组成键值对并排序，前后组合secret并生成md5签名，并判断与传入的sign是否相等
func (w *Input) CheckAPISign(secret string) Error {
	if f, ok := w.Ext["__checkAPIAuth__"].(func(string) Error); ok {
		return f(secret)
	}
	return NewError(ERR_NOT_EXTENDED, "不支持APIAuth")

}

//CheckInput 检查输入参数
func (w *Input) CheckInput(names ...string) error {
	for _, v := range names {
		if r, err := w.Input.Get(v); err != nil || r == "" {
			err = fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

//CheckArgs 检查args参数
func (w *Input) CheckArgs(names ...string) error {
	for _, v := range names {
		if r, ok := w.Args[v]; !ok || r == "" {
			err := fmt.Errorf("args配置中缺少参数:%s", v)
			return err
		}
	}
	return nil
}
func (w *Input) Has(names ...string) error {
	for _, name := range names {
		if _, err := w.Input.Get(name); err != nil {
			return fmt.Errorf("不包含:%s", name)
		}
	}

	return nil
}
func (w *Input) Get(name string) (string, error) {
	return w.Input.Get(name)
}

//GetString 从input获取字符串数据
func (w *Input) GetString(name string, p ...string) string {
	t, err := w.Input.Get(name)
	if err == nil {
		return t
	}
	if len(p) > 0 {
		return p[0]
	}
	return ""
}

//GetJWTBody 获取jwt存储数据
func (w *Input) GetJWTBody() interface{} {
	return w.Ext["__jwt_"]
}

//GetInt 从input中获取int数字
func (w *Input) GetInt(name string) (int, error) {
	value, err := w.Get(name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("input.%s的值不是有效的int值", name)
		return 0, err
	}
	return v, nil
}

//GetInt64 从input中获取int64数字
func (w *Input) GetInt64(name string) (int64, error) {
	value, err := w.Get(name)
	if err != nil {
		return 0, err
	}
	iv, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return iv, nil
	}
	return 0, err
}

//DecodeBody2Input 根据编码格式解码body参数，并更新input参数
func (w *Input) DecodeBody2Input(encoding ...string) error {
	body, err := w.DecodeBody(encoding...)
	if err != nil {
		return err
	}
	qString, err := utility.GetMapWithQuery(body)
	if err != nil {
		return err
	}
	for k, v := range qString {
		w.Input.Set(k, v)
	}
	return nil
}

//DecodeBody 解码body参数
func (w *Input) DecodeBody(encoding ...string) (string, error) {
	if len(encoding) == 0 {
		return w.Body, nil
	}
	if fun, ok := w.Ext["__func_body_get_"].(func(ch string) (string, error)); ok {
		return fun(encoding[0])
	}
	return "", fmt.Errorf("无法根据%s格式转换数据", encoding[0])
}

func (w *Input) getVarParam() (func(c string, n string) (string, error), error) {
	funcVar := w.Ext["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_var_get_传入类型错误")
}

//GetVarParam 获取var参数值，需提供在ext中提供__func_var_get_
func (w *Input) GetVarParam(tpName string, name string) (string, error) {
	f, err := w.getVarParam()
	if err != nil {
		return "", err
	}
	return f(tpName, name)
}

//GetArgsValue 获取arg.value值
func (w *Input) GetArgsValue(name string, d ...string) string {
	v, _ := w.GetArgsByName(name)
	if v == "" && len(d) > 0 {
		return d[0]
	}
	return v
}

//GetArgsByName 获取arg的参数
func (w *Input) GetArgsByName(name string) (string, error) {
	db, ok := w.Args[name]
	if db == "" || !ok {
		return "", fmt.Errorf("args配置错误，缺少:%s参数:%v", name, w.Args)
	}
	return db, nil
}
func (w *Input) GetArgsInt(name string, v ...int) int {
	r, err := w.GetArgsIntValue(name)
	if err == nil {
		return r
	}
	if len(v) > 0 {
		return v[0]
	}
	return 0
}
func (w *Input) GetArgsInt64(name string, v ...int64) int64 {
	r, err := w.GetArgsInt64Value(name)
	if err == nil {
		return r
	}
	if len(v) > 0 {
		return v[0]
	}
	return 0
}

//GetArgsIntValue 从args中获取int数字
func (w *Input) GetArgsIntValue(name string) (int, error) {
	value, err := w.GetArgsByName(name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("arg.%s的值不是有效的int值", name)
		return 0, err
	}
	return v, nil
}

//GetArgsInt64Value 从args中获取int64数字
func (w *Input) GetArgsInt64Value(name string) (int64, error) {
	value, err := w.GetArgsByName(name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return v, nil
	}
	return 0, err
}

//GetArgsFloat64Value 从args中获取float64数字
func (w *Input) GetArgsFloat64Value(name string) (float64, error) {
	value, err := w.GetArgsByName(name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		err = fmt.Errorf("arg.%s的值不是有效的float值", name)
		return 0, err
	}
	return v, nil
}

//GetVarParamByArgsName 根据args参数名获取var参数的值
func (w *Input) GetVarParamByArgsName(tpName string, argsName string) (string, error) {
	name, err := w.GetArgsByName(argsName)
	if err != nil {
		return "", err
	}
	return w.GetVarParam(tpName, name)
}
