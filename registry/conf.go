package registry

import "fmt"

import "github.com/qxnw/lib4go/registry"

//Registry 注册中心接口
type Registry interface {
	Exists(path string) (bool, error)
	WatchChildren(path string) (data chan registry.ChildrenWatcher, err error)
	WatchValue(path string) (data chan registry.ValueWatcher, err error)
	GetChildren(path string) (data []string, version int32, err error)
	GetValue(path string) (data []byte, version int32, err error)
	CreatePersistentNode(path string, data string) (err error)
	CreateTempNode(path string, data string) (err error)
	CreateSeqNode(path string, data string) (rpath string, err error)
}

//Conf 配置提供从配置文件中读取参数的方法
type Conf interface {
	GetVersion() int32
	String(key string, def ...string) string      //support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
	Strings(key string, def ...[]string) []string //get string slice
	Int(key string, def ...int) (int, error)
	Bool(key string, def ...bool) (bool, error)
	GetSection(section string) (Conf, error)
	GetNode(section string) (Conf, error)
	GetSections(section string) (cs []Conf, err error)
	Len() int
}

const (
	ADD = iota + 1
	CHANGE
	DEL
)

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
