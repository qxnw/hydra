package creator

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type IMainBinder interface {
	SetMainConf(s string)
	SetSubConf(n string, s string)
	Scan(platName string, mainConf string) error
	NeedScanCount() int
	GetNodeConf() map[string]string
}

//MainBinder 主配置绑定
type MainBinder struct {
	mainConf                   string                       //系统主配置
	subConf                    map[string]string            //子系统配置
	mainParamsForInput         []string                     //主配置参数，用于用户输入
	subParamsForInput          map[string][]string          //子系统参数,用于用户输入
	mainConfParamsForTranslate map[string]string            //主配置参数，用于参数翻译
	subConfParamsForTranslate  map[string]map[string]string //子系统参数,用于参数翻译
	rmainConf                  string                       //翻译后的主配置
	rsubConf                   map[string]string            //翻译后的子系统配置
}

//NewMainBinder 构建主配置绑定
func NewMainBinder() *MainBinder {
	return &MainBinder{
		subConf:                    make(map[string]string),
		mainParamsForInput:         make([]string, 0, 2),
		subParamsForInput:          make(map[string][]string),
		mainConfParamsForTranslate: make(map[string]string),
		subConfParamsForTranslate:  make(map[string]map[string]string),
		rsubConf:                   make(map[string]string),
	}
}

//SetMainConf 设置主配置内容
func (c *MainBinder) SetMainConf(s string) {
	c.mainConf = s
	c.mainParamsForInput = getParams(s)
}

//SetSubConf 设置子配置内容
func (c *MainBinder) SetSubConf(n string, s string) {
	c.subConf[n] = s
	params := getParams(s)
	if len(params) > 0 {
		c.subParamsForInput[n] = params
	}
}

//NeedScanCount 待输入个数
func (c *MainBinder) NeedScanCount() int {
	return len(c.mainParamsForInput) + len(c.subParamsForInput)
}

//Scan 绑定参数
func (c *MainBinder) Scan(platName string, mainConf string) error {
	for _, p := range c.mainParamsForInput {
		fmt.Printf("请输入:%s中%s的值:", mainConf, p)
		var value string
		fmt.Scan(&value)
		c.mainConfParamsForTranslate[p] = value

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
	c.rmainConf = translate(c.mainConf, c.mainConfParamsForTranslate)
	for k, v := range c.subConf {
		c.rsubConf[filepath.Join(mainConf, k)] = translate(v, c.subConfParamsForTranslate[k])
	}
	return nil
}

//GetNodeConf 获取节点配置
func (c *MainBinder) GetNodeConf() map[string]string {
	nmap := make(map[string]string)
	nmap["."] = c.rmainConf
	for k, v := range c.rsubConf {
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
