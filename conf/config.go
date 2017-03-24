package conf

import "fmt"

//Config 配置提供从配置文件中读取参数的方法
type Config interface {
	String(key string, def ...string) string      //support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
	Strings(key string, def ...[]string) []string //get string slice
	Int(key string, def ...int) (int, error)
	Bool(key string, def ...bool) (bool, error)
	GetSection(section string) (Config, error)
	GetSections(section string) (cs []Config, err error)
	Len() int
}

//ConfWatcher 配置文件监控器
type ConfWatcher interface {
	Notify() chan []Config
	Get() []Config
}

//ConfigAdapter 定义配置文件转换方法
type ConfigAdapter interface {
	Parse(key ...string) (Config, error)
	ParseData(data []byte) (Config, error)
}

var adapters = make(map[string]ConfigAdapter)

//Register 注册配置文件适配器
func Register(name string, adapter ConfigAdapter) {
	if adapter == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

//NewConfig 根据适配器名称及参数返回配置处理器
func NewConfig(adapterName string, args ...string) (Config, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("config: unknown adaptername %q (forgotten import?)", adapterName)
	}
	return adapter.Parse(args...)
}

//NewConfigData 根据配置数据生成配置处理器
func NewConfigData(adapterName string, data []byte) (Config, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("config: unknown adaptername %q (forgotten import?)", adapterName)
	}
	return adapter.ParseData(data)
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
