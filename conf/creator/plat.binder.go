package creator

import (
	"fmt"
	"path/filepath"
)

type IPlatBinder interface {
	SetVarConf(t string, s string, v string)
	Scan(platName string) error
	NeedScanCount() int
	GetNodeConf() map[string]string
}

//PlatBinder 平台配置绑定
type PlatBinder struct {
	varConf                   map[string]string            //var环境参数配置
	varParamsForInput         map[string][]string          //环境参数，用于用户输入
	varConfParamsForTranslate map[string]map[string]string //环境参数，用于参数翻译
	rvarConf                  map[string]string            //翻译后的环境参数配置
}

//NewPlatBinder 平台绑定
func NewPlatBinder() *PlatBinder {
	return &PlatBinder{
		varConf: make(map[string]string),
	}
}

//SetVarConf 设置var配置内容
func (c *PlatBinder) SetVarConf(t string, s string, v string) {
	c.varConf[filepath.Join(t, s)] = v
	params := getParams(s)
	if len(params) > 0 {
		c.varParamsForInput[filepath.Join(t, s)] = params
	}
}

//NeedScanCount 待输入个数
func (c *PlatBinder) NeedScanCount() int {
	return len(c.varParamsForInput)
}

//Scan 绑定参数
func (c *PlatBinder) Scan(platName string) error {
	for n, ps := range c.varParamsForInput {
		c.varConfParamsForTranslate[n] = make(map[string]string)
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:", filepath.Join(platName, n), p)
			var value string
			fmt.Scan(&value)
			c.varConfParamsForTranslate[n][p] = value
		}
	}
	for k, v := range c.varConf {
		c.rvarConf[filepath.Join(platName, k)] = translate(v, c.varConfParamsForTranslate[k])
	}
	return nil
}

//GetNodeConf 获取节点配置
func (c *PlatBinder) GetNodeConf() map[string]string {
	return c.rvarConf
}
