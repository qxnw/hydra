package hydra

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry/service"
	"github.com/qxnw/lib4go/logger"

	"strings"

	"time"

	"github.com/qxnw/hydra/server"
)

var (
	modeStandalone      = "standalone"
	modeCluster         = "cluster"
	errServerIsExist    = errors.New("服务已存在")
	errServerIsNotExist = errors.New("服务不存在")
)

//Server hydra的单个服务器示例
type Server struct {
	domain                  string
	runMode                 string
	engine                  engine.IEngine
	server                  server.IHydraServer
	extModes                string
	engineNames             []string
	serviceRegistry         service.IService
	crurrentRegistryAddress []string
	crossRegistryAddress    []string
	registry                string
	localServices           []string
	remoteServices          []string
	serverName              string
	serverType              string
	runTime                 time.Time
	address                 string
	logger                  *logger.Logger
}

//NewHydraServer 初始化服务器
func NewHydraServer(domain string, runMode string, registry string, crurrentRegistryAddress []string, crossRegistryAddress []string, logger *logger.Logger) *Server {
	return &Server{
		domain:                  domain,
		runMode:                 runMode,
		registry:                registry,
		remoteServices:          make([]string, 0, 16),
		localServices:           make([]string, 0, 16),
		crurrentRegistryAddress: crurrentRegistryAddress,
		crossRegistryAddress:    crossRegistryAddress,
		engine:                  engine.NewStandardEngine(),
		logger:                  logger,
	}
}

//Start 启用服务器
func (h *Server) Start(cnf conf.Conf) (err error) {
	h.serverName = cnf.String("name")
	h.serverType = cnf.String("type")
	if strings.EqualFold(cnf.String("status"), server.ST_STOP) {
		return fmt.Errorf("server未启动:%s(%s) 配置为:%s", cnf.String("name"), h.serverType, cnf.String("status"))
	}

	h.extModes = cnf.String("extModes")
	h.engineNames = cnf.Strings("extModes", []string{})
	h.engineNames = append(h.engineNames, []string{"go", "rpc"}...)
	h.serviceRegistry, err = service.NewRegister(h.runMode, h.domain, h.serverName, h.logger, h.crurrentRegistryAddress, h.crossRegistryAddress)
	if err != nil {
		return fmt.Errorf("register初始化失败 mode:%s,domain:%s(err:%v)", h.serverType, h.domain, err)
	}

	// 启动执行引擎
	h.localServices, err = h.engine.Start(h.domain, h.serverName, h.serverType, h.registry, h.logger, h.engineNames...)
	if err != nil {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(%s)(err:%v)", h.domain, h.serverName, h.serverType, err)
	}
	if !server.IsDebug && strings.EqualFold(h.serverType, server.SRV_TP_RPC) && len(h.localServices) == 0 {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(%s)(err:engine中未找到任何服务)", h.domain, h.serverName, h.serverType)
	}

	//构建服务器
	h.server, err = server.NewServer(h.serverType, h.engine, h.serviceRegistry, cnf)
	if err != nil {
		return fmt.Errorf("server启动失败:%s(%s)(err:%v)", h.serverName, h.serverType, err)
	}
	err = h.server.Start()
	if err != nil {
		return fmt.Errorf("server启动失败:%s(%s)(err:%v)", h.serverName, h.serverType, err)
	}
	h.address = h.server.GetAddress()
	h.runTime = time.Now()

	//注册服务列表
	if strings.EqualFold(h.serverType, server.SRV_TP_RPC) {
		for _, v := range h.localServices {
			path, err := h.serviceRegistry.RegisterTempNode(v, strings.Replace(h.server.GetAddress(), "//", "", -1), h.server.GetAddress())
			if err != nil {
				return err
			}
			h.remoteServices = append(h.remoteServices, path)
		}
	}
	return nil
}

//Notify 配置发生变化通知服务器变更
func (h *Server) Notify(cnf conf.Conf) error {
	return h.server.Notify(cnf)
}

//GetStatus 获取当前服务状态
func (h *Server) GetStatus() string {
	return h.server.GetStatus()
}

//EngineHasChange 引擎配置发生变化
func (h *Server) EngineHasChange(p string) bool {
	return p != h.extModes
}

//Shutdown 关闭服务器
func (h *Server) Shutdown() {
	if h.serviceRegistry != nil {
		for _, v := range h.remoteServices {
			h.serviceRegistry.Unregister(v)
		}
	}
	h.engine.Close()
	h.server.Shutdown()
}
