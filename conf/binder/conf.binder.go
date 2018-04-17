package binder

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type IBinder interface {
	SetMainConf(s string)
	SetSubConf(n string, s string)
	SetVarConf(t string, s string, v string)
	Bind(platName string, mainConf string) error
	GetNodeConf() map[string]string
}

//Binder 创建配置
type Binder struct {
	platName     string
	mainConfPath string            //系统主配置路径
	mainConf     string            //系统主配置
	subConf      map[string]string //子系统配置
	varConf      map[string]string //var环境参数配置

	rmainConf string            //翻译后的主配置
	rsubConf  map[string]string //翻译后的子系统配置
	rvarConf  map[string]string //翻译后的环境参数配置

	mainParamsForInput         map[string][]string          //主配置参数，用于用户输入
	subParamsForInput          map[string][]string          //子系统参数,用于用户输入
	varParamsForInput          map[string][]string          //环境参数，用于用户输入
	mainConfParamsForTranslate map[string]string            //主配置参数，用于参数翻译
	subConfParamsForTranslate  map[string]map[string]string //子系统参数,用于参数翻译
	varConfParamsForTranslate  map[string]map[string]string //环境参数，用于参数翻译
}

func NewBinder() *Binder {
	return &Binder{
		subConf:                    make(map[string]string),
		varConf:                    make(map[string]string),
		rsubConf:                   make(map[string]string),
		rvarConf:                   make(map[string]string),
		mainParamsForInput:         make(map[string][]string),
		subParamsForInput:          make(map[string][]string),
		varParamsForInput:          make(map[string][]string),
		mainConfParamsForTranslate: make(map[string]string),
		subConfParamsForTranslate:  make(map[string]map[string]string),
		varConfParamsForTranslate:  make(map[string]map[string]string),
	}
}

//SetMainConf 设置主配置内容
func (c *Binder) SetMainConf(s string) {
	c.mainConf = s
	params := getParams(s)
	if len(params) > 0 {
		c.mainParamsForInput[""] = params
	}
}

//SetSubConf 设置子配置内容
func (c *Binder) SetSubConf(n string, s string) {
	c.subConf[n] = s
	params := getParams(s)
	if len(params) > 0 {
		c.subParamsForInput[n] = params
	}
}

//SetVarConf 设置var配置内容
func (c *Binder) SetVarConf(t string, s string, v string) {
	c.varConf[filepath.Join(t, s)] = v
	params := getParams(s)
	if len(params) > 0 {
		c.varParamsForInput[filepath.Join(t, s)] = params
	}
}

//Bind 绑定参数
func (c *Binder) Bind(platName string, mainConf string) error {
	if len(c.mainParamsForInput) > 0 || len(c.subParamsForInput) > 0 || len(c.varParamsForInput) > 0 {
		var index string
		fmt.Print("当前应用程序启动需要一些关键的参数才能启动，是否立即设置这些参数(yes|NO):")
		fmt.Scan(&index)
		if index != "y" && index != "Y" && index != "yes" && index != "YES" {
			return nil
		}
	}
	for n, ps := range c.mainParamsForInput {
		for _, p := range ps {
			fmt.Printf("请输入:%s%s中%s的值:", mainConf, n, p)
			var value string
			fmt.Scan(&value)
			c.mainConfParamsForTranslate[p] = value
		}
	}
	for n, ps := range c.subParamsForInput {
		c.subConfParamsForTranslate[n] = make(map[string]string)
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:", filepath.Join(mainConf, n), p)
			var value string
			fmt.Scan(&value)
			c.subConfParamsForTranslate[n][p] = value
		}
	}
	for n, ps := range c.varParamsForInput {
		c.varConfParamsForTranslate[n] = make(map[string]string)
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:", filepath.Join(platName, n), p)
			var value string
			fmt.Scan(&value)
			c.varConfParamsForTranslate[n][p] = value
		}
	}
	c.rmainConf = translate(mainConf, c.mainConfParamsForTranslate)
	for k, v := range c.subConf {
		c.rsubConf[filepath.Join(mainConf, k)] = translate(v, c.subConfParamsForTranslate[k])
	}
	for k, v := range c.varConf {
		c.rvarConf[filepath.Join(platName, k)] = translate(v, c.varConfParamsForTranslate[k])
	}
	return nil
}

//GetNodeConf 获取节点配置
func (c *Binder) GetNodeConf() map[string]string {
	nmap := make(map[string]string)
	nmap[c.mainConfPath] = c.rmainConf
	for k, v := range c.rsubConf {
		nmap[k] = v
	}
	for k, v := range c.varConf {
		nmap[k] = v
	}
	return nmap
}

//getParams 翻译带有@变量的字符串
func getParams(format string) []string {
	brackets, _ := regexp.Compile(`\{@\w+\}`)
	p1 := brackets.FindAllString(format, -1)
	brackets, _ = regexp.Compile(`@\w+`)
	p2 := brackets.FindAllString(format, -1)
	r := make([]string, 0, len(p1)+len(p2))
	for _, v := range p1 {
		r = append(r, v[2:len(v)-1])
	}
	for _, v := range p2 {
		r = append(r, v[1:])
	}
	return r
}

//translate 翻译带有@变量的字符串
func translate(format string, data map[string]string) string {
	brackets, _ := regexp.Compile(`\{@\w+\}`)
	result := brackets.ReplaceAllStringFunc(format, func(s string) string {
		key := s[2 : len(s)-1]
		return data[key]
	})
	word, _ := regexp.Compile(`@\w+`)
	result = word.ReplaceAllStringFunc(result, func(s string) string {
		key := s[1:]
		return data[key]
	})
	return result
}
