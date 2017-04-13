package conf

import "fmt"

//WatchServices /api，web，rpc，job
var WatchServices = []string{"api", "web", "rpc", "job"}

type Updater struct {
	Conf Conf
	Op   int
}

//ConfWatcher 配置文件监控器
type ConfWatcher interface {
	Start() error
	Notify() chan *Updater
	Close() error
}

//ConfResolver 定义配置文件转换方法
type ConfResolver interface {
	Resolve(adapter string, domain string, tag string, args ...string) (ConfWatcher, error)
}

var confResolvers = make(map[string]ConfResolver)

//Register 注册配置文件适配器
func Register(name string, resolver ConfResolver) {
	if resolver == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := confResolvers[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	confResolvers[name] = resolver
}

//NewWatcher 根据适配器名称及参数返回配置处理器
func NewWatcher(adapter string, domain string, tag string, args ...string) (ConfWatcher, error) {
	resolver, ok := confResolvers[adapter]
	if !ok {
		return nil, fmt.Errorf("config: unknown adapter name %q (forgotten import?)", adapter)
	}
	return resolver.Resolve(adapter, domain, tag, args...)
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
