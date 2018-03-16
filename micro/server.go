package micro

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/engines"
	registry "github.com/qxnw/hydra/registry.v2"
	"github.com/qxnw/hydra/registry.v2/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"

	"time"
)

var (
	errServerIsExist    = errors.New("服务已存在")
	errServerIsNotExist = errors.New("服务不存在")
)

//Server hydra的单个服务器示例
type server struct {
	registry     registry.IRegistry
	cnf          conf.IServerConf
	registryAddr string

	startTime time.Time
	logger    *logger.Logger
	engine    engines.IServiceEngine
	server    servers.IRegistryServer
}

//newServer 初始化服务器
func newServer(cnf conf.IServerConf, registry registry.IRegistry) *server {
	return &server{
		registry: registry,
		cnf:      cnf,
	}
}

//Start 启用服务器
func (h *server) Start() (err error) {
	h.logger = logger.New(h.cnf.GetPlatName())
	h.logger.Infof("开始启动:%s", h.cnf.GetPlatName())

	// 启动执行引擎
	h.engine, err = engines.NewServiceEngine(h.cnf.GetPlatName(), h.cnf.GetSysName(), h.cnf.GetServerType(), h.registryAddr, h.logger, h.cnf.GetStrings("engines", "go", "rpc")...)
	if err != nil {
		return fmt.Errorf("%s:engine启动失败%v", h.cnf.GetPlatName(), err)
	}

	//构建服务器
	h.server, err = servers.NewRegistryServer(h.cnf.GetServerType(), h.engine, h.cnf, h.logger)
	if err != nil {
		return fmt.Errorf("server初始化失败:%s", h.cnf.GetServerName())
	}
	err = h.server.Start()
	if err != nil {
		return fmt.Errorf("server启动失败:%s", h.cnf.GetServerName())
	}
	h.startTime = time.Now()
	return nil
}

//Notify 配置发生变化通知服务器变更
func (h *server) Notify(cnf conf.IServerConf) error {
	return h.server.Notify(cnf)
}

//GetStatus 获取当前服务状态
func (h *server) GetStatus() string {
	return h.server.GetStatus()
}

//Shutdown 关闭服务器
func (h *server) Shutdown() {
	if h.server != nil {
		h.server.Shutdown()
	}
	if h.engine != nil {
		h.engine.Close()
	}
}
