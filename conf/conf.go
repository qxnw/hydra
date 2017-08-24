package conf

import (
	"fmt"

	"github.com/qxnw/lib4go/transform"
)

//Conf 基本JSON的配置管理器
type Conf interface {
	GetVersion() int32
	Has(key string) bool
	String(key string, def ...string) string      //support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
	Strings(key string, def ...[]string) []string //get string slice
	Int(key string, def ...int) (int, error)
	Bool(key string, def ...bool) (bool, error)
	GetData() map[string]interface{}
	GetSection(section string) (Conf, error)
	GetIMap(section string) (map[string]interface{}, error)
	GetSMap(section string) (map[string]string, error)
	GetRawNodeWithValue(value string, enableCache ...bool) (r []byte, err error)
	GetNodeWithSectionValue(sectionValue string, enableCache ...bool) (r Conf, err error)
	GetNodeWithSectionName(sectionName string, defValue ...string) (Conf, error)
	GetSections(section string) (cs []Conf, err error)
	GetSectionString(section string) (r string, err error)
	GetArray(key string) (r []interface{}, err error)
	GetContent() string
	Each(f func(key string))
	Translate(format string) string
	Set(string, string)
	Len() int
	Append(t transform.ITransformGetter)
}

//ParseBool 将字符串转换为bool值
func ParseBool(val interface{}) (value bool, err error) {
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
