package conf

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IConf interface {
	GetVersion() int32
	GetString(key string, def ...string) (r string)
	GetStrings(key string, def ...string) (r []string)
	GetArray(key string, def ...interface{}) (r []interface{})
	GetBool(key string, def ...bool) (r bool, err error)
	GeJSON(section string) (r []byte, version int32, err error)
	GetSection(section string) (c *JSONConf, err error)
}

//JSONConf json配置文件
type JSONConf struct {
	raw     json.RawMessage
	version int32
	data    map[string]interface{}
}

//NewJSONConfByMap 根据map初始化对象
func NewJSONConfByMap(data map[string]interface{}, version int32) (c *JSONConf, err error) {
	c = &JSONConf{
		data:    data,
		version: version,
	}
	return c, nil
}

//NewJSONConf 初始化JsonConf
func NewJSONConf(message []byte, version int32) (c *JSONConf, err error) {
	c = &JSONConf{
		raw:     json.RawMessage(message),
		version: version,
	}
	if err = json.Unmarshal(message, &c.data); err != nil {
		return
	}
	return c, nil
}

//GetVersion 获取当前配置的版本号
func (j *JSONConf) GetVersion() int32 {
	return j.version
}

//GetString 获取字符串
func (j *JSONConf) GetString(key string, def ...string) (r string) {
	if val, ok := j.data[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		default:
			return fmt.Sprint(val)
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

//GetStrings 获取字符串数组
func (j *JSONConf) GetStrings(key string, def ...string) (r []string) {
	if r = strings.Split(j.GetString(key), ";"); len(r) > 0 {
		return r
	}
	if len(def) > 0 {
		return def
	}
	return nil
}

//GetArray 获取数组对象
func (j *JSONConf) GetArray(key string, def ...interface{}) (r []interface{}) {
	if _, ok := j.data[key]; !ok {
		if len(def) > 0 {
			return def
		}
		return nil
	}
	if r, ok := j.data[key].([]interface{}); ok {
		return r
	}
	return nil
}

//GetBool 获取bool类型值
func (j *JSONConf) GetBool(key string, def ...bool) (r bool, err error) {
	if val := j.GetString(key); val != "" {
		return parseBool(val)
	}
	if len(def) > 0 {
		return def[0], nil
	}
	return false, fmt.Errorf("%s不是bool类型值", key)
}

//GeJSON 获取section原始JSON串
func (j *JSONConf) GeJSON(section string) (r []byte, version int32, err error) {
	if v, ok := j.data[section]; !ok || v == nil {
		err = fmt.Errorf("节点:%s不存在或值为空", section)
		return
	}
	val := j.data[section]
	buffer, err := json.Marshal(val)
	if err != nil {
		return nil, 0, err
	}
	return buffer, j.version, nil
}

//GetSection 指定节点名称获取JSONConf
func (j *JSONConf) GetSection(section string) (c *JSONConf, err error) {
	if v, ok := j.data[section]; !ok || v == nil {
		err = fmt.Errorf("节点:%s不存在或值为空", section)
		return
	}
	if data, ok := j.data[section].(map[string]interface{}); ok {
		return NewJSONConfByMap(data, j.version)
	}
	err = fmt.Errorf("节点:%s不是有效的json对象", section)
	return
}

//ParseBool 将字符串转换为bool值
func parseBool(val interface{}) (value bool, err error) {
	if val != nil {
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			switch v {
			case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "Y", "y", "ON", "on", "On":
				return true, nil
			case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "N", "n", "OFF", "off", "Off":
				return false, nil
			}
		case int8, int32, int64:
			strV := fmt.Sprintf("%s", v)
			if strV == "1" {
				return true, nil
			} else if strV == "0" {
				return false, nil
			}
		case float64:
			if v == 1 {
				return true, nil
			} else if v == 0 {
				return false, nil
			}
		}
		return false, fmt.Errorf("parsing %q: invalid syntax", val)
	}
	return false, fmt.Errorf("parsing <nil>: invalid syntax")
}