package context

import (
	"fmt"
	"strconv"
)

type inputParams struct {
	data IData
}

//Check 检查是否包含指定的参数
func (i *inputParams) Check(names ...string) error {
	for _, v := range names {
		if r, err := i.Get(v); err != nil || r == "" {
			return fmt.Errorf("字段:%v值不能为空", v)
		}
	}
	return nil
}
func (i *inputParams) Each(f func(string, string)) {
	i.data.Each(f)
}
func (i *inputParams) Get(name string, p ...string) (string, error) {
	return i.data.Get(name)
}
func (i *inputParams) GetString(name string, p ...string) string {
	v, err := i.Get(name)
	if err == nil {
		return v
	}
	if len(p) > 0 {
		return p[0]
	}
	return ""
}

//GetInt 获取int数字
func (i *inputParams) GetInt(name string, p ...int) int {
	value, err := i.Get(name)
	var v int
	if err == nil {
		v, err = strconv.Atoi(value)
	}
	if err == nil {
		return v
	}
	if len(p) > 0 {
		return p[0]
	}
	return 0
}

//GetInt64 获取int64数字
func (i *inputParams) GetInt64(name string, p ...int64) int64 {
	value, err := i.Get(name)
	var v int64
	if err == nil {
		v, err = strconv.ParseInt(value, 10, 64)
	}
	if err == nil {
		return v
	}
	if len(p) > 0 {
		return p[0]
	}
	return 0
}

//GetFloat64 获取float64数字
func (i *inputParams) GetFloat64(name string, p ...float64) float64 {
	value, err := i.Get(name)
	var v float64
	if err == nil {
		v, err = strconv.ParseFloat(value, 64)
	}
	if err == nil {
		return v
	}
	if len(p) > 0 {
		return p[0]
	}
	return 0
}
