package component

import (
	"fmt"

	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

//IComponentCache Component Cache
type IComponentCache interface {
	GetDefaultCache() (c cache.ICache, err error)
	GetCache(name string) (c cache.ICache, err error)
	Close() error
}

//StandardCache cache
type StandardCache struct {
	IContainer
	name     string
	cacheMap cmap.ConcurrentMap
}

//NewStandardCache 创建cache
func NewStandardCache(c IContainer, name ...string) *StandardCache {
	if len(name) > 0 {
		return &StandardCache{IContainer: c, name: name[0], cacheMap: cmap.New(2)}
	}
	return &StandardCache{IContainer: c, name: "cache", cacheMap: cmap.New(2)}
}

//GetDefaultCache 获取默然配置cache
func (s *StandardCache) GetDefaultCache() (c cache.ICache, err error) {
	return s.GetCache(s.name)
}

//GetCache 获取缓存操作对象
func (s *StandardCache) GetCache(name string) (c cache.ICache, err error) {
	return s.GetCacheByTypeName("cache", name)
}

//GetCacheByTypeName 根据类型获取缓存数据
func (s *StandardCache) GetCacheByTypeName(tpName string, name string) (c cache.ICache, err error) {
	_, cached, err := s.cacheMap.SetIfAbsentCb(name, func(input ...interface{}) (c interface{}, err error) {
		name := input[0].(string)
		conf, err := s.IContainer.GetVarConf(tpName, name)
		if err != nil {
			return nil, err
		}
		c, err = cache.NewCache(conf.GetString("server"), string(conf.GetRaw()))
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

//Close 关闭缓存连接
func (s *StandardCache) Close() error {
	s.cacheMap.RemoveIterCb(func(k string, v interface{}) bool {
		v.(cache.ICache).Close()
		return true
	})
	return nil
}
