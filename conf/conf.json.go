package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/registry"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
)

//JSONConf json配置文件
type JSONConf struct {
	data     map[string]interface{}
	cache    cmap.ConcurrentMap
	Content  string
	registry registry.Registry
	version  int32
	*transform.Transform
}

//NewJSONConfWithJson 根据JSON初始化conf对象
func NewJSONConfWithJson(c string, version int32, rgst registry.Registry) (r *JSONConf, err error) {
	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(c), &m)
	if err != nil {
		return
	}
	return &JSONConf{
		Content:   c,
		data:      m,
		cache:     cmap.New(8),
		Transform: transform.NewMaps(m),
		version:   version,
		registry:  rgst,
	}, nil
}

//NewJSONConfWithEmpty 初始化空的conf对象
func NewJSONConfWithEmpty() *JSONConf {
	return NewJSONConfWithHandle(make(map[string]interface{}), 0, nil)
}

//NewJSONConfWithHandle 根据map和动态获取函数构建
func NewJSONConfWithHandle(m map[string]interface{}, version int32, registry registry.Registry) *JSONConf {

	return &JSONConf{
		data:      m,
		cache:     cmap.New(8),
		Transform: transform.NewMaps(m),
		version:   version,
		registry:  registry,
	}
}

//GetData 获取原始数据
func (j *JSONConf) GetData() map[string]interface{} {
	r, _ := jsons.Unmarshal([]byte(j.Content))
	return r
}

//GetContent 获取输入JSON原串
func (j *JSONConf) GetContent() string {
	return j.Content
}

//Len 获取参数个数
func (j *JSONConf) Len() int {
	return len(j.data)
}

//GetVersion 获取当前配置的版本号
func (j *JSONConf) GetVersion() int32 {
	return j.version
}

//Set 设置参数值
func (j *JSONConf) Set(key string, value string) {
	if _, ok := j.data[key]; ok {
		return
	}
	j.Transform.Set(key, value)
}

//String 获取字符串，已缓存则从缓存中获取
func (j *JSONConf) String(key string, def ...string) (r string) {
	if v, err := j.Transform.Get(key); err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

//GetArray 获取数组列表
func (j *JSONConf) GetArray(key string) (r []interface{}, err error) {
	d, ok := j.data[key]
	if !ok {
		err = fmt.Errorf("不包含数据:%s", key)
		return
	}
	if r, ok := d.([]interface{}); ok {
		return r, nil
	}
	err = fmt.Errorf("不包含数据:%s", key)
	return
}

//Strings 获取字符串数组，原字符串以“;”号分隔
func (j *JSONConf) Strings(key string, def ...[]string) (r []string) {
	stringVal := j.String(key)
	if stringVal != "" {
		r = strings.Split(j.String(key), ";")
		return
	}
	if len(def) > 0 {
		return def[0]
	}
	return []string{}
}

//Bool 获取BOOL参数
func (j *JSONConf) Bool(key string, def ...bool) (r bool, err error) {
	if val, ok := j.data[key]; ok {
		return ParseBool(val)
	}
	if len(def) > 0 {
		return def[0], nil
	}
	err = fmt.Errorf("not exist key: %q", key)
	return
}

//Has 检查当前配置是否包含指定key
func (j *JSONConf) Has(key string) bool {
	if strings.HasPrefix(key, "#") {
		if j.registry == nil {
			return false
		}
		path := j.TranslateAll(key, true)
		if b, err := j.registry.Exists(path[1:]); err == nil {
			return b
		}
	}
	if _, err := j.Transform.Get(key); err == nil {
		return true
	}
	return false
}

//Int 获取整数值
func (j *JSONConf) Int(key string, def ...int) (r int, err error) {
	val := j.data[key]
	if val != nil {
		if v, ok := val.(int); ok {
			r = int(v)
			return
		} else if v, ok := val.(float64); ok {
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

//GetRawNodeWithValue 获取配置的路径节点的原始值，节点必须以#开头
func (j *JSONConf) GetRawNodeWithValue(nodeValue string, enableCache ...bool) (r []byte, err error) {
	if j.registry == nil {
		return nil, fmt.Errorf("获取%s未指定数据获取方式:registry is nil", nodeValue)
	}
	if !strings.HasPrefix(nodeValue, "#") {
		return nil, fmt.Errorf("该节点的值不允许使用GetNode方法获取：%s", nodeValue)
	}
	rx := j.TranslateAll(nodeValue[1:], true)
	r, _, err = j.registry.GetValue(rx)
	if err != nil {
		return
	}
	return
}

//GetNodeWithSectionValue 获取节点的配置值，节点必须以#开头
func (j *JSONConf) GetNodeWithSectionValue(nodeValue string, enableCache ...bool) (r Conf, err error) {
	if j.registry == nil {
		return nil, fmt.Errorf("获取%s未指定数据获取方式:registry is nil", nodeValue)
	}
	if !strings.HasPrefix(nodeValue, "#") {
		return nil, fmt.Errorf("该节点的值不允许使用GetNode方法获取：%s", nodeValue)
	}
	nkey := "_node_with_value_" + nodeValue
	if value, ok := j.cache.Get(nkey); ok {
		r = value.(Conf)
		return
	}
	path := j.TranslateAll(nodeValue[1:], true)
	buff, v, err := j.registry.GetValue(path)
	if err != nil {
		r, _ = NewJSONConfWithJson("{}", 0, j.registry)
		j.cache.Set(nkey, r)
		return
	}
	r, err = NewJSONConfWithJson(string(buff), v, j.registry)
	if err != nil {
		return
	}
	j.cache.Set(nkey, r)
	return
}

//GetNodeWithSectionName 获取子节点的配置数据
func (j *JSONConf) GetNodeWithSectionName(sectionName string, defValue ...string) (r Conf, err error) {
	if j.registry == nil {
		return nil, fmt.Errorf("获取%s未指定数据获取方式:registry is nil", sectionName)
	}
	value := j.String(sectionName)
	if value == "" && len(defValue) == 0 {
		err = fmt.Errorf("节点:%s的值为空", sectionName)
		return
	}

	if value == "" {
		value = defValue[0]
	}
	r, err = j.GetNodeWithSectionValue(value, false)
	if err != nil {
		return
	}
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

//Each 循化每个节点值
func (j *JSONConf) Each(f func(key string)) {
	for k := range j.data {
		f(k)
	}
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

//GetSectionString 获取section原始JSON串
func (j *JSONConf) GetSectionString(section string) (r string, err error) {
	//nkey := "_section_string_" + section
	//if value, ok := j.cache.Get(nkey); ok {
	//return value.(string), nil
	//}
	val := j.data[section]
	if val != nil {
		buffer, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		r = j.TranslateAll(jsons.Escape(string(buffer)), false)
		//	j.cache.Set(nkey, r)
		return r, nil
	}
	err = errors.New("not exist section:" + section)
	return
}

//GetSection 获取块节点
func (j *JSONConf) GetSection(section string) (r Conf, err error) {
	val := j.data[section]
	if val != nil {
		if v, ok := val.(map[string]interface{}); ok {
			r = NewJSONConfWithHandle(v, j.version, j.registry)
			return
		}
	}
	err = errors.New("not exist section:" + section)
	return
}

//GetSections 获取配置列表
func (j *JSONConf) GetSections(section string) (cs []Conf, err error) {
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
				cs = append(cs, NewJSONConfWithHandle(nmap, j.version, j.registry))
			}
		}
		return
	}
	err = errors.New("not exist section:" + section)
	return
}
