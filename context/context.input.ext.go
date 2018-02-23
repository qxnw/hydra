package context

import (
	"fmt"
	"strings"

	"github.com/qxnw/lib4go/utility"
)

type extParams struct {
	ext map[string]interface{}
}

func (w *extParams) Get(name string) (interface{}, bool) {
	v, ok := w.ext[name]
	return v, ok
}
func (w *extParams) GetMethod() string {
	if m, ok := w.ext["__method_"].(string); ok {
		return m
	}
	return ""
}
func (w *extParams) GetHeader() map[string]string {
	if h, ok := w.ext["__header_"].(map[string]string); ok {
		return h
	}
	header := make(map[string]string)
	if h, ok := w.ext["__header_"].(map[string][]string); ok {
		for k, v := range h {
			header[k] = strings.Join(v, ",")
		}
	}
	return header
}

//GetSharding 获取任务分片信息(分片索引[从1开始]，分片总数)
func (w *extParams) GetSharding() (int, int) {
	v, ok := w.ext["__get_sharding_index_"]
	if !ok {
		return 0, 0
	}
	if f, ok := v.(func() (int, int)); ok {
		return f()
	}
	return 0, 0
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

/*
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
}*/

//GetJWTBody 获取jwt存储数据
func (w *extParams) GetJWTBody() interface{} {
	return w.ext["__jwt_"]
}

//GetUUID
func (w *extParams) GetUUID() string {
	return w.ext["__hydra_sid_"].(string)
}
