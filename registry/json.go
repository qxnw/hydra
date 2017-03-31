package registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/lib4go/transform"
)

//JSONConf json配置文件
type JSONConf struct {
	data  map[string]interface{}
	cache map[string]interface{}
	*transform.Transform
}

//NewJSONConf 构建JSON配置文件
func NewJSONConf(m map[string]interface{}) *JSONConf {
	return &JSONConf{
		data:      m,
		cache:     make(map[string]interface{}),
		Transform: transform.NewMaps(m),
	}
}

//Len 参数个数
func (j *JSONConf) Len() int {
	return len(j.data)
}

//String 获取字符串
func (j *JSONConf) String(key string, def ...string) (r string) {
	if value, ok := j.cache[key]; ok {
		r = value.(string)
		return
	}
	val := j.data[key]
	if val != nil {
		if v, ok := val.(string); ok {
			r = j.Translate(v)
			j.cache[key] = r
			return r
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

//Strings 获取字符串数组，原字符串以“;”号分隔
func (j *JSONConf) Strings(key string, def ...[]string) (r []string) {
	if value, ok := j.cache[key]; ok {
		r = value.([]string)
		return
	}
	stringVal := j.String(key)
	if stringVal != "" {
		return strings.Split(j.String(key), ";")
	}
	if len(def) > 0 {
		return def[0]
	}
	return []string{}
}

//Bool 获取BOOL参数
func (j *JSONConf) Bool(key string, def ...bool) (r bool, err error) {
	if value, ok := j.cache[key]; ok {
		r = value.(bool)
		return
	}
	val := j.data[key]
	if val != nil {
		return ParseBool(val)
	}
	if len(def) > 0 {
		return def[0], nil
	}
	err = fmt.Errorf("not exist key: %q", key)
	return
}

//Int 获取整数值
func (j *JSONConf) Int(key string, def ...int) (r int, err error) {
	if value, ok := j.cache[key]; ok {
		r = value.(int)
		return
	}
	val := j.data[key]
	if val != nil {
		if v, ok := val.(int); ok {
			r = int(v)
			return
		}
		err = errors.New("not int value")
		return
	}
	if len(def) > 0 {
		r = def[0]
		return
	}
	err = errors.New("not exist key:" + key)
	return
}

//GetSection 获取块节点
func (j *JSONConf) GetSection(section string) (r Conf, err error) {
	if value, ok := j.cache[section]; ok {
		return value.(Conf), nil
	}
	val := j.data[section]
	if val != nil {
		if v, ok := val.(map[string]interface{}); ok {
			r = NewJSONConf(v)
			j.cache[section] = r
			return
		}
	}
	err = errors.New("not exist section:" + section)
	return
}

//GetSections 获取配置列表
func (j *JSONConf) GetSections(section string) (cs []Conf, err error) {
	if value, ok := j.cache[section]; ok {
		return value.([]Conf), nil
	}
	cs = make([]Conf, 0, 0)
	val := j.data[section]
	if val != nil {
		if v, ok := val.([]interface{}); ok {
			for _, value := range v {
				nmap := make(map[string]interface{})
				if m, ok := value.(map[string]interface{}); ok {
					for x, y := range m {
						nmap[x] = y
					}
				}
				for x, y := range j.data {
					if _, ok := nmap[x]; !ok {
						nmap[x] = y
					}
				}
				cs = append(cs, NewJSONConf(nmap))
			}
		}
		j.cache[section] = cs
		return
	}
	err = errors.New("not exist section:" + section)
	return
}
