package component

import (
	"fmt"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
)
//IComponentCache Component Cache
type IComponentCache interface{
	GetDefaultCache() (c cache.ICache, err error)
	GetCache(name string) (c cache.ICache, err error)
}
//StandardCache cache
type StandardCache struct {
	engine.IContainer
	name string
}

//NewStandardCache 创建cache
func NewStandardCache(c engine.IContainer, name ...string) *StandardCache {
	if len(name) > 0 {
		return &StandardCache{IContainer: c, name: name[0]}
	}
	return &StandardCache{IContainer: c, name: "cache"}
}

//GetDefaultCache 获取默然配置cache
func (s *StandardCache) GetDefaultCache() (c cache.ICache, err error) {
	return s.GetCache(s.name)
}

//GetCache 获取缓存操作对象
func (s *StandardCache) GetCache(name string) (c cache.ICache, err error) {
	_, cached, err := cacheMap.SetIfAbsentCb(name, func(input ...interface{}) (c interface{}, err error) {
		name := input[0].(string)
		conf, err := s.IContainer.GetVarParam("cache", name)
		if err != nil {
			return nil, err
		}
		configMap, err := jsons.Unmarshal([]byte(conf))
		if err != nil {
			return nil, err
		}
		server, ok := configMap["server"]
		if !ok {
			err = fmt.Errorf("cache[%s]配置文件错误，未包含server节点:%s", name, conf)
			return nil, err
		}
		c, err = cache.NewCache(server.(string), conf)
		if err != nil {
			return nil, err
		}
		return
	}, name)
	if err != nil {
		err = fmt.Errorf("初始化cache失败:%v", err)
		return
	}
	c = cached.(cache.ICache)
	return
}

var cacheMap cmap.ConcurrentMap

func init() {
	cacheMap = cmap.New(2)
}
