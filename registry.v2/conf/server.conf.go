package conf

import (
	"fmt"
	"path/filepath"
	"strings"

	registry "github.com/qxnw/hydra/registry.v2"
)

//IServerConf 服务器配置
type IServerConf interface {
	GetPlatName() string
	GetSysName() string
	GetServerType() string
	GetClusterName() string
	GetServerName() string
	IConf
	GetSystemConf(name string) *JSONConf
	GetVarConf(tp string, name string) *JSONConf
}

//ServerConf 服务器配置信息
type ServerConf struct {
	*JSONConf
	platName       string
	sysName        string
	serverType     string
	clusterName    string
	mainConfpath   string
	varConfPath    string
	childNodeConfs map[string]*JSONConf
	varNodeConfs   map[string]*JSONConf
	registry       registry.IRegistry
}

//NewServerConf 构建服务器配置缓存
func NewServerConf(mainConfpath string, mainConfRaw []byte, mainConfVersion int32, rgst registry.IRegistry) (s *ServerConf, err error) {
	sections := strings.Split(strings.Trim(mainConfpath, "/"), "/")
	if len(sections) != 5 {
		err = fmt.Errorf("conf配置文件格式错误，格式:/platName/sysName/serverType/clusterName/conf")
		return
	}
	s = &ServerConf{
		mainConfpath:   mainConfpath,
		platName:       sections[0],
		sysName:        sections[1],
		serverType:     sections[2],
		clusterName:    sections[3],
		varConfPath:    filepath.Join(sections[0], "var"),
		registry:       rgst,
		childNodeConfs: make(map[string]*JSONConf),
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
		wather, err := NewJSONConf(data, version)
		if err != nil {
			err = fmt.Errorf("%s配置有误:%v", childConfPath, err)
			return err
		}
		c.childNodeConfs[p] = wather
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
			wather, err := NewJSONConf(data, version)
			if err != nil {
				err = fmt.Errorf("%s配置有误:%v", nodePath, err)
				return err
			}
			c.varNodeConfs[filepath.Join(p, node)] = wather
		}
	}
	return nil
}

//GetSystemConf 指定配置文件名称，获取系统配置信息
func (c *ServerConf) GetSystemConf(name string) *JSONConf {
	if v, ok := c.childNodeConfs[name]; ok {
		return v
	}
	v, _ := NewJSONConfByMap(make(map[string]interface{}), 0)
	return v
}

//GetVarConf 指定配置文件名称，获取var配置信息
func (c *ServerConf) GetVarConf(tp string, name string) *JSONConf {
	if v, ok := c.varNodeConfs[filepath.Join(tp, name)]; ok {
		return v
	}
	return nil
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
