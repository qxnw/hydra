package conf

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/qxnw/hydra/registry"
)

//ErrNoSetting 未配置
var ErrNoSetting = errors.New("未配置")

type ISystemConf interface {
	GetPlatName() string
	GetSysName() string
	GetServerType() string
	GetClusterName() string
	GetServerName() string
}
type IMainConf interface {
	IConf
	IsStop() bool
	GetSubObject(name string, v interface{}) (int32, error)
	GetSubConf(name string) (*JSONConf, error)
	HasSubConf(name ...string) bool
}
type IVarConf interface {
	GetVarConf(tp string, name string) (*JSONConf, error)
	GetVarObject(tp string, name string, v interface{}) (int32, error)
	HasVarConf(tp string, name string) bool
	GetVarConfClone() map[string]JSONConf
	SetVarConf(map[string]JSONConf)
}

//IServerConf 服务器配置
type IServerConf interface {
	ISystemConf
	IMainConf
	IVarConf
}

//ServerConf 服务器配置信息
type ServerConf struct {
	*JSONConf
	platName     string
	sysName      string
	serverType   string
	clusterName  string
	mainConfpath string
	varConfPath  string
	subNodeConfs map[string]JSONConf
	varNodeConfs map[string]JSONConf
	registry     registry.IRegistry
	varLock      sync.RWMutex
}

//NewServerConf 构建服务器配置缓存
func NewServerConf(mainConfpath string, mainConfRaw []byte, mainConfVersion int32, rgst registry.IRegistry) (s *ServerConf, err error) {
	sections := strings.Split(strings.Trim(mainConfpath, "/"), "/")
	if len(sections) != 5 {
		err = fmt.Errorf("conf配置文件格式错误，格式:/platName/sysName/serverType/clusterName/conf")
		return
	}
	s = &ServerConf{
		mainConfpath: mainConfpath,
		platName:     sections[0],
		sysName:      sections[1],
		serverType:   sections[2],
		clusterName:  sections[3],
		varConfPath:  filepath.Join("/", sections[0], "var"),
		registry:     rgst,
		subNodeConfs: make(map[string]JSONConf),
		varNodeConfs: make(map[string]JSONConf),
	}
	//初始化主配置
	if s.JSONConf, err = NewJSONConf(mainConfRaw, mainConfVersion); err != nil {
		err = fmt.Errorf("%s配置有误:%v", mainConfpath, err)
		return nil, err
	}
	if err = s.loadChildNodeConf(); err != nil {
		return
	}
	if err = s.loadVarNodeConf(); err != nil {
		return
	}
	return s, nil
}

//初始化子节点配置
func (c *ServerConf) loadChildNodeConf() error {
	paths, _, err := c.registry.GetChildren(c.mainConfpath)
	if err != nil {
		return err
	}
	for _, p := range paths {
		childConfPath := filepath.Join(c.mainConfpath, p)
		data, version, err := c.registry.GetValue(childConfPath)
		if err != nil {
			return err
		}
		childConf, err := NewJSONConf(data, version)
		if err != nil {
			err = fmt.Errorf("%s配置有误:%v", childConfPath, err)
			return err
		}
		c.subNodeConfs[p] = *childConf
	}
	return nil
}

//初始化子节点配置
func (c *ServerConf) loadVarNodeConf() error {
	//获取第一级目录
	varfirstNodes, _, err := c.registry.GetChildren(c.varConfPath)
	if err != nil {
		return err
	}

	for _, p := range varfirstNodes {
		//获取第二级目录
		firstNodePath := filepath.Join(c.varConfPath, p)
		varSecondChildren, _, err := c.registry.GetChildren(firstNodePath)
		if err != nil {
			return err
		}

		//获取二级目录的值
		for _, node := range varSecondChildren {
			nodePath := filepath.Join(firstNodePath, node)
			data, version, err := c.registry.GetValue(nodePath)
			if err != nil {
				return err
			}
			varConf, err := NewJSONConf(data, version)
			if err != nil {
				err = fmt.Errorf("%s配置有误:%v", nodePath, err)
				return err
			}
			c.varNodeConfs[filepath.Join(p, node)] = *varConf
		}
	}
	return nil
}

//IsStop 当前服务是否已停止
func (c *ServerConf) IsStop() bool {
	return c.GetString("status") == "stop"
}

//GetSubObject 获取子系统配置
func (c *ServerConf) GetSubObject(name string, v interface{}) (int32, error) {
	conf, err := c.GetSubConf(name)
	if err != nil {
		return 0, err
	}
	if err := conf.Unmarshal(v); err != nil {
		return 0, err
	}
	return conf.version, nil
}

//GetSubConf 指定配置文件名称，获取系统配置信息
func (c *ServerConf) GetSubConf(name string) (*JSONConf, error) {
	if v, ok := c.subNodeConfs[name]; ok {
		return &v, nil
	}
	return nil, ErrNoSetting
}

//HasSubConf 是否存在子级配置
func (c *ServerConf) HasSubConf(names ...string) bool {
	for _, name := range names {
		_, ok := c.subNodeConfs[name]
		if ok {
			return true
		}
	}
	return false
}

//GetVarConf 指定配置文件名称，获取var配置信息
func (c *ServerConf) GetVarConf(tp string, name string) (*JSONConf, error) {
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	if v, ok := c.varNodeConfs[filepath.Join(tp, name)]; ok {
		return &v, nil
	}
	return nil, ErrNoSetting
}

//GetVarConfClone 获取var配置拷贝
func (c *ServerConf) GetVarConfClone() map[string]JSONConf {
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	data := make(map[string]JSONConf)
	for k, v := range c.varNodeConfs {
		data[k] = v
	}
	return data
}

//SetVarConf 获取var配置参数
func (c *ServerConf) SetVarConf(data map[string]JSONConf) {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	c.varNodeConfs = data
}

//GetVarObject 指定配置文件名称，获取var配置信息
func (c *ServerConf) GetVarObject(tp string, name string, v interface{}) (int32, error) {
	conf, err := c.GetVarConf(tp, name)
	if err != nil {
		return 0, err
	}
	if err := conf.Unmarshal(&v); err != nil {
		return 0, err
	}
	return conf.version, nil
}

//HasVarConf 是否存在子级配置
func (c *ServerConf) HasVarConf(tp string, name string) bool {
	_, ok := c.varNodeConfs[filepath.Join(tp, name)]
	return ok
}

//GetPlatName 获取平台名称
func (c *ServerConf) GetPlatName() string {
	return c.platName
}

//GetSysName 获取系统名称
func (c *ServerConf) GetSysName() string {
	return c.sysName
}

//GetServerType 获取服务器类型
func (c *ServerConf) GetServerType() string {
	return c.serverType
}

//GetClusterName 获取集群名称
func (c *ServerConf) GetClusterName() string {
	return c.clusterName
}

//GetServerName 获取服务器名称
func (c *ServerConf) GetServerName() string {
	return fmt.Sprintf("%s.%s(%s,%s)", c.sysName, c.platName, c.serverType, c.clusterName)
}
