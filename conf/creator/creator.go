package creator

import (
	"path/filepath"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//Creator 配置文件创建器
type Creator struct {
	registry    registry.IRegistry
	logger      *logger.Logger
	binder      IBinder
	platName    string
	systemName  string
	serverTypes []string
	clusterName string
}

//NewCreator 配置文件创建器
func NewCreator(platName string, systemName string, serverTypes []string, clusterName string, binder IBinder, rgst registry.IRegistry, logger *logger.Logger) (w *Creator) {
	w = &Creator{
		platName:    platName,
		systemName:  systemName,
		serverTypes: serverTypes,
		clusterName: clusterName,
		registry:    rgst,
		logger:      logger,
		binder:      binder,
	}
	return
}

//Start 绑定输入参数并创建配置文件
func (c *Creator) Start() (err error) {
	if err := c.binder.Scan(c.platName, c.systemName, c.serverTypes, c.clusterName); err != nil {
		return err
	}
	for _, tp := range c.serverTypes {
		p := filepath.Join("/", c.platName, c.systemName, tp, c.clusterName, "conf")
		conf := c.binder.GetMainConf(tp)
		data, ok := conf["."]
		if ok {
			if err = c.createMainConf(p, data); err != nil {
				return err
			}
		} else {
			if err = c.createMainConf(p, "{}"); err != nil {
				return err
			}
		}
		for k, v := range conf {
			if k != "." {
				if err = c.registry.CreatePersistentNode(k, v); err != nil {
					return err
				}
			}
		}

	}
	//创建平台配置
	platConfs := c.binder.GetPlatConf()
	for k, v := range platConfs {
		if err = c.registry.CreatePersistentNode(k, v); err != nil {
			return err
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
