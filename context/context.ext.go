package context

import (
	"errors"
	"fmt"
	"strconv"
)

//GetVarParam 获取var参数值，需提供在ext中提供__func_var_get_
func (c *Context) GetVarParam(tpName string, name string) (string, error) {
	func_var := c.GetExt()["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		s, err := f(tpName, name)
		if err != nil {
			err = fmt.Errorf("无法通过获取到参数/@domain/var/%s/%s的值", tpName, name)
			return "", err
		}
		return s, nil
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}

//CheckArgs 检查必须参数
func (c *Context) CheckArgs(names ...string) error {
	argsMap := c.GetArgs()
	for _, name := range names {
		db, ok := argsMap[name]
		if db == "" || !ok {
			return fmt.Errorf("args配置错误，缺少:%s参数:%v", name, c.GetArgs())
		}
	}
	return nil
}

//GetArgValue 获取arg.value值
func (c *Context) GetArgValue(name string, d ...string) string {
	v, _ := c.GetArgByName(name)
	if v == "" && len(d) > 0 {
		return d[0]
	}
	return v
}

//GetArgByName 获取arg的参数
func (c *Context) GetArgByName(name string) (string, error) {
	argsMap := c.GetArgs()
	db, ok := argsMap[name]
	if db == "" || !ok {
		return "", fmt.Errorf("args配置错误，缺少:%s参数:%v", name, c.GetArgs())
	}
	return db, nil
}

//GetArgIntValue 从args中获取int数字
func (c *Context) GetArgIntValue(name string) (int, error) {
	value, err := c.GetArgByName(name)
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

//GetArgFloat64Value 从args中获取float64数字
func (c *Context) GetArgFloat64Value(name string) (float64, error) {
	value, err := c.GetArgByName(name)
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
func (c *Context) GetVarParamByArgsName(tpName string, argsName string) (string, error) {
	name, err := c.GetArgByName(argsName)
	if err != nil {
		return "", err
	}
	return c.GetVarParam(tpName, name)
}
