package component

import (
	"encoding/json"
	"fmt"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
)
//IComponentConf Component config
type IComponentConf interface{
	GetConf(conf interface{}) (c interface{}, err error)
}
//StandardConf conf
type StandardConf struct {
	engine.IContainer
}

//NewStandardConf 创建conf
func NewStandardConf(c engine.IContainer) *StandardConf {
	return &StandardConf{IContainer: c}
}

//GetConf 获取置信息
func (s *StandardConf) GetConf(conf interface{}) (c interface{}, err error) {
	name := s.IContainer.GetServerName()
	_, v, err := confCache.SetIfAbsentCb(name, func(input ...interface{}) (interface{}, error) {
		name := input[0].(string)
		content, err := s.IContainer.GetVarParam("conf", name)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(content), conf)
		if err != nil {
			err = fmt.Errorf("conf配置文件错误:%v", err)
			return nil, err
		}
		return conf, nil
	}, name)
	if err != nil {
		return nil, err
	}
	return v, nil
}

var confCache cmap.ConcurrentMap

func init() {
	confCache = cmap.New(2)
}
