package context

import (
	"errors"
	"fmt"
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
func (c *Context) GetArgValue(name string) string {
	v, _ := c.GetArgByName(name)
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

//GetVarParamByArgsName 根据args参数名获取var参数的值
func (c *Context) GetVarParamByArgsName(tpName string, argsName string) (string, error) {
	name, err := c.GetArgByName(argsName)
	if err != nil {
		return "", err
	}
	return c.GetVarParam(tpName, name)
}
