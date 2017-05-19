package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/transform"
)

//JSONConf json配置文件
type JSONConf struct {
	data    map[string]interface{}
	cache   cmap.ConcurrentMap
	Content string
	handle  func(path string) (Conf, error)
	version int32
	*transform.Transform
}

func NewJSONConfWithJson(c string, version int32, handle func(path string) (Conf, error)) (r *JSONConf, err error) {
	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(c), &m)
	if err != nil {
		return
	}
	m["now"] = time.Now().Format("2006/01/02 15:04:05")
	return &JSONConf{
		Content:   c,
		data:      m,
		cache:     cmap.New(),
		Transform: transform.NewMaps(m),
		version:   version,
		handle:    handle,
	}, nil
}

func NewJSONConfWithEmpty() *JSONConf {
	return NewJSONConfWithHandle(make(map[string]interface{}), 0, func(string) (Conf, error) {
		return NewJSONConfWithEmpty(), nil
	})
}

//NewJSONConfWithHandle 根据map和动态获取函数构建
func NewJSONConfWithHandle(m map[string]interface{}, version int32, handle func(path string) (Conf, error)) *JSONConf {
	m["now"] = time.Now().Format("2006/01/02 15:04:05")
	return &JSONConf{
		data:      m,
		cache:     cmap.New(),
		Transform: transform.NewMaps(m),
		version:   version,
		handle:    handle,
	}
}
func (j *JSONConf) GetContent() string {
	return j.Content
}

//Len 参数个数
func (j *JSONConf) Len() int {
	return len(j.data)
}

//GetVersion 获取当前配置的版本号
func (j *JSONConf) GetVersion() int32 {
	return j.version
}

//Set 设置参数值
func (j *JSONConf) Set(key string, value string) {
	j.data[key] = value
	j.Transform.Set(key, value)
}

//String 获取字符串
func (j *JSONConf) String(key string, def ...string) (r string) {
	nkey := "_s_" + key
	if value, ok := j.cache.Get(nkey); ok {
		r = value.(string)
		return
	}
	val := j.data[key]
	if val != nil {
		if v, ok := val.(string); ok {
			r = j.TranslateAll(v, false)
			j.cache.Set(nkey, r)
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
	nkey := "_strs_" + key
	if value, ok := j.cache.Get(nkey); ok {
		r = value.([]string)
		return
	}
	stringVal := j.String(key)
	if stringVal != "" {
		r = strings.Split(j.String(key), ";")
		j.cache.Set(nkey, r)
		return
	}
	if len(def) > 0 {
		return def[0]
	}
	return []string{}
}

//Bool 获取BOOL参数
func (j *JSONConf) Bool(key string, def ...bool) (r bool, err error) {
	nkey := "_b_" + key
	if value, ok := j.cache.Get(nkey); ok {
		r = value.(bool)
		return
	}
	val := j.data[key]
	if val != nil {
		r, err = ParseBool(val)
		if err != nil {
			return
		}
		j.cache.Set(nkey, r)
		return
	}
	if len(def) > 0 {
		return def[0], nil
	}
	err = fmt.Errorf("not exist key: %q", key)
	return
}
func (j *JSONConf) Has(key string) bool {
	if _, ok := j.data[key]; ok {
		return true
	}
	return false
}

//Int 获取整数值
func (j *JSONConf) Int(key string, def ...int) (r int, err error) {
	nkey := "_int_" + key
	if value, ok := j.cache.Get(nkey); ok {
		r = value.(int)
		return
	}
	val := j.data[key]
	if val != nil {
		if v, ok := val.(int); ok {
			r = int(v)
			j.cache.Set(nkey, r)
			return
		} else if v, ok := val.(float64); ok {
			r = int(v)
			j.cache.Set(nkey, r)
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

//GetNodeWithValue 获取节点值
func (j *JSONConf) GetNodeWithValue(value string, enableCache ...bool) (r Conf, err error) {
	if j.handle == nil {
		return nil, errors.New("未指定NODE获取方式")
	}

	ec := true
	if len(enableCache) > 0 {
		ec = enableCache[0]
	}
	nkey := "_node_" + value
	if value, ok := j.cache.Get(nkey); ok && ec {
		r = value.(Conf)
		return
	}
	if !strings.HasPrefix(value, "#") {
		return nil, fmt.Errorf("该节点的值不允许使用GetNode方法获取：%s", value)
	}
	r, err = j.handle(j.TranslateAll(value[1:], true))
	if err != nil {
		return
	}
	j.cache.Set(nkey, r)
	return
}

//GetNodeWithSection 获取节点值
func (j *JSONConf) GetNodeWithSection(section string, enableCache ...bool) (r Conf, err error) {
	if j.handle == nil {
		return nil, errors.New("未指定NODE获取方式")
	}
	ec := true
	if len(enableCache) > 0 {
		ec = enableCache[0]
	}
	nkey := "_node_section_" + section
	if value, ok := j.cache.Get(nkey); ok && ec {
		r = value.(Conf)
		return
	}
	r, err = j.GetNodeWithValue(j.String(section), enableCache...)
	if err != nil {
		return
	}
	j.cache.Set(nkey, r)
	return
}

//GetIMap 获取map数据
func (j *JSONConf) GetIMap(section string) (map[string]interface{}, error) {
	val := j.data[section]
	if val != nil {
		if v, ok := val.(map[string]interface{}); ok {
			return v, nil
		}
	}
	return nil, nil
}

//GetSMap 获取map数据
func (j *JSONConf) GetSMap(section string) (map[string]string, error) {
	data := make(map[string]string)
	val := j.data[section]
	if val != nil {
		if v, ok := val.(map[string]interface{}); ok {
			for k, a := range v {
				data[k] = j.TranslateAll(fmt.Sprintf("%s", a), false)
			}
		}
	}
	return data, nil
}

//GetSection 获取块节点
func (j *JSONConf) GetSection(section string) (r Conf, err error) {
	nkey := "_section_" + section
	if value, ok := j.cache.Get(nkey); ok {
		return value.(Conf), nil
	}
	val := j.data[section]
	if val != nil {
		if v, ok := val.(map[string]interface{}); ok {
			r = NewJSONConfWithHandle(v, j.version, j.handle)
			j.cache.Set(nkey, r)
			return
		}
	}
	err = errors.New("not exist section:" + section)
	return
}

//GetSections 获取配置列表
func (j *JSONConf) GetSections(section string) (cs []Conf, err error) {
	nkey := "_section_" + section
	if value, ok := j.cache.Get(nkey); ok {
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
				cs = append(cs, NewJSONConfWithHandle(nmap, j.version, j.handle))
			}
		}
		j.cache.Set(nkey, cs)
		return
	}
	err = errors.New("not exist section:" + section)
	return
}
