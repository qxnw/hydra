package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/transform"
)

type JsonConfAdapter struct {
}

// Parse creates a new Config and parses the file configuration from the named file.
func (j *JsonConfAdapter) Parse(args ...string) (conf.Config, error) {
	if len(args) == 0 {
		return nil, errors.New("输入参数不能为空")
	}
	name := args[0]
	filename := fmt.Sprintf("../%s/conf/conf.json", strings.Trim(name, "/"))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return j.ParseData(content)
}

func (j *JsonConfAdapter) ParseData(data []byte) (conf.Config, error) {
	c := make(map[string]interface{})
	err := json.Unmarshal(data, &c)
	if err != nil {
		var wrappingArray []interface{}
		err2 := json.Unmarshal(data, &wrappingArray)
		if err2 != nil {
			return nil, err
		}
	}
	return NewJSONConf(c), nil
}

//JSONConfig json配置文件
type JSONConf struct {
	data  map[string]interface{}
	cache map[string]interface{}
	*transform.Transform
}

//NewJSONConfig 构建JSON配置文件
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
		return conf.ParseBool(val)
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
func (j *JSONConf) GetSection(section string) (r conf.Config, err error) {
	if value, ok := j.cache[section]; ok {
		return value.(conf.Config), nil
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
func (j *JSONConf) GetSections(section string) (cs []conf.Config, err error) {
	if value, ok := j.cache[section]; ok {
		return value.([]conf.Config), nil
	}
	cs = make([]conf.Config, 0, 0)
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
func init() {
	conf.Register("json", &JsonConfAdapter{})
}
