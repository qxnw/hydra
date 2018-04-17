package conf

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

//Creator 创建配置
type Creator struct {
	platName     string
	mainConfPath string
	MainConf     string
	SubConf      map[string]string
	VarConf      map[string]string

	rMainConf string
	rSubConf  map[string]string
	rVarConf  map[string]string

	mainParams     map[string][]string
	subParams      map[string][]string
	varParams      map[string][]string
	mainConfParams map[string]string
	subConfParams  map[string]map[string]string
	varConfParams  map[string]map[string]string
}

func NewCreator(mainConfPath string) *Creator {
	return &Creator{
		mainConfPath:   mainConfPath,
		platName:       strings.Split(strings.TrimLeft(mainConfPath, string(filepath.Separator)), string(filepath.Separator))[0],
		SubConf:        make(map[string]string),
		VarConf:        make(map[string]string),
		rSubConf:       make(map[string]string),
		rVarConf:       make(map[string]string),
		mainParams:     make(map[string][]string),
		subParams:      make(map[string][]string),
		varParams:      make(map[string][]string),
		mainConfParams: make(map[string]string),
		subConfParams:  make(map[string]map[string]string),
		varConfParams:  make(map[string]map[string]string),
	}
}

//SetMainConf 设置主配置内容
func (c *Creator) SetMainConf(s string) {
	c.MainConf = s
	params := getParams(s)
	if len(params) > 0 {
		c.mainParams[c.mainConfPath] = params
	}
}

//SetSubConf 设置子配置内容
func (c *Creator) SetSubConf(n string, s string) {
	c.SubConf[n] = s
	params := getParams(s)
	if len(params) > 0 {
		c.subParams[filepath.Join(c.mainConfPath, n)] = params
	}
}

//SetVarConf 设置var配置内容
func (c *Creator) SetVarConf(t string, s string, v string) {
	c.VarConf[filepath.Join(t, s)] = v
	params := getParams(s)
	if len(params) > 0 {
		c.varParams[filepath.Join(c.platName, t, s)] = params
	}
}

//Bind 绑定参数
func (c *Creator) Bind() error {
	if len(c.mainParams) > 0 || len(c.subParams) > 0 || len(c.varParams) > 0 {
		var index string
		fmt.Println("当前应用程序启动需要一些关键的参数才能启动，是否立即设置这些参数(yes|NO)")
		fmt.Scan(&index)
		if index != "y" && index != "Y" && index != "yes" && index != "YES" {
			return nil
		}
	}
	for n, ps := range c.mainParams {
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:\n", n, p)
			var value string
			fmt.Scan(&value)
			c.mainConfParams[p] = value
		}
	}
	for n, ps := range c.subParams {
		c.subConfParams[n] = make(map[string]string)
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:\n", n, p)
			var value string
			fmt.Scan(&value)
			c.subConfParams[n][p] = value
		}
	}
	for n, ps := range c.varParams {
		c.varConfParams[n] = make(map[string]string)
		for _, p := range ps {
			fmt.Printf("请输入:%s中%s的值:\n", n, p)
			var value string
			fmt.Scan(&value)
			c.varConfParams[n][p] = value
		}
	}
	c.rMainConf = translate(c.MainConf, c.mainConfParams)
	for k, v := range c.SubConf {
		c.rSubConf[k] = translate(v, c.subConfParams[k])
	}
	for k, v := range c.VarConf {
		c.rVarConf[k] = translate(v, c.varConfParams[k])
	}
	return nil
}

//GetNodeConf 获取节点配置
func (c *Creator) GetNodeConf() map[string]string {
	nmap := make(map[string]string)
	nmap[c.mainConfPath] = c.rMainConf
	for k, v := range c.rSubConf {
		nmap[k] = v
	}
	for k, v := range c.VarConf {
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
