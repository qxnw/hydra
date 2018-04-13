package component

import (
	"fmt"
	"path/filepath"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

//IComponentVarObject Component Cache
type IComponentVarObject interface {
	GetObject(tpName string, name string) (c interface{}, err error)
	SaveObject(tpName string, name string, f func(c conf.IConf) (interface{}, error)) (bool, interface{}, error)
	Close() error
}

//VarObjectCache cache
type VarObjectCache struct {
	IContainer
	cacheMap  cmap.ConcurrentMap
	closeList []CloseHandler
}

//NewVarObjectCache 创建cache
func NewVarObjectCache(c IContainer) *VarObjectCache {
	return &VarObjectCache{IContainer: c, cacheMap: cmap.New(2), closeList: make([]CloseHandler, 0, 1)}
}

//GetObject 根据类型获取缓存数据
func (s *VarObjectCache) GetObject(tpName string, name string) (c interface{}, err error) {
	cacheConf, err := s.IContainer.GetVarConf(tpName, name)
	if err != nil {
		return nil, fmt.Errorf("%s %v", filepath.Join("/", s.GetPlatName(), "var", tpName, name), err)
	}
	key := fmt.Sprintf("%s/%s:%d", tpName, name, cacheConf.GetVersion())
	c, ok := s.cacheMap.Get(key)
	if !ok {
		err = fmt.Errorf("缓存对象未创建:%s", filepath.Join("/", s.GetPlatName(), "var", tpName, name))
		return
	}
	return c, nil
}

//SaveObject 缓存对象
func (s *VarObjectCache) SaveObject(tpName string, name string, f func(c conf.IConf) (interface{}, error)) (bool, interface{}, error) {
	cacheConf, err := s.IContainer.GetVarConf(tpName, name)
	if err != nil {
		return false, nil, fmt.Errorf("%s %v", filepath.Join("/", s.GetPlatName(), "var", tpName, name), err)
	}
	key := fmt.Sprintf("%s/%s:%d", tpName, name, cacheConf.GetVersion())
	ok, ch, err := s.cacheMap.SetIfAbsentCb(key, func(input ...interface{}) (c interface{}, err error) {
		c, err = f(cacheConf)
		if err != nil {
			return nil, err
		}
		switch v := c.(type) {
		case CloseHandler:
			s.closeList = append(s.closeList, v)
		}
		return c, nil
	})
	if err != nil {
		err = fmt.Errorf("创建对象失败:%s,err:%v", string(cacheConf.GetRaw()), err)
		return ok, nil, err
	}
	return ok, ch, err
}

//Close 关闭缓存连接
func (s *VarObjectCache) Close() error {
	s.cacheMap.Clear()
	for _, f := range s.closeList {
		f.Close()
	}
	return nil
}
