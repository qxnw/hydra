package binder

import (
	"path/filepath"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//Creator 配置文件创建器
type Creator struct {
	registry    registry.IRegistry
	logger      *logger.Logger
	binder      *ServerBinder
	platName    string
	systemName  string
	serverTypes []string
	clusterName string
	paths       []string
}

//NewCreator 配置文件创建器
func NewCreator(platName string, systemName string, serverTypes []string, clusterName string, binder *ServerBinder, rgst registry.IRegistry, logger *logger.Logger) (w *Creator) {
	w = &Creator{
		platName:    platName,
		systemName:  systemName,
		serverTypes: serverTypes,
		clusterName: clusterName,
		registry:    rgst,
		logger:      logger,
		binder:      binder,
	}
	w.paths = make([]string, 0, len(serverTypes))
	for _, tp := range serverTypes {
		w.paths = append(w.paths, filepath.Join("/", platName, systemName, tp, clusterName, "conf"))
	}
	return
}

//Start 绑定输入参数并创建配置文件
func (c *Creator) Start() (err error) {

	if err := c.binder.Bind(c.platName, c.systemName, c.serverTypes, c.clusterName); err != nil {
		return err
	}

	//创建主配置
	for i, p := range c.paths {
		tp := c.serverTypes[i]
		data, ok := c.binder.GetConfs(tp)[p]
		if ok {
			err = c.createMainConf(p, data)
		} else {
			err = c.createMainConf(p, "{}")
		}
		if err != nil {
			return err
		}
	}

	//创建其它配置
	for _, tp := range c.serverTypes {
		data := c.binder.GetConfs(tp)
		for k, v := range data {
			e := false
			for _, p := range c.paths {
				if p == k {
					e = true
					break
				}
			}
			if !e {
				if err = c.registry.CreatePersistentNode(k, v); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
func (c *Creator) createMainConf(path string, data string) error {
	extPath := ""
	if !c.registry.CanWirteDataInDir() {
		extPath = ".init"
	}
	if data == "" {
		data = "{}"
	}
	rpath := filepath.Join(path, extPath)
	b, err := c.registry.Exists(rpath)
	if err != nil {
		return err
	}
	if !b {
		if err := c.registry.CreatePersistentNode(rpath, data); err != nil {
			return err
		}
	}
	return nil
}
