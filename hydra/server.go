package hydra

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/registry/service"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/logger"

	"strings"

	"time"
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
	tag                     string
	engine                  engines.IServiceEngine
	server                  servers.IRegistryServer
	engines                 string
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
func NewHydraServer(domain string, tag string, runMode string, registry string, crurrentRegistryAddress []string, crossRegistryAddress []string) *Server {
	return &Server{
		domain:                  domain,
		tag:                     tag,
		runMode:                 runMode,
		registry:                registry,
		localServices:           make([]string, 0, 16),
		crurrentRegistryAddress: crurrentRegistryAddress,
		crossRegistryAddress:    crossRegistryAddress,
	}
}

//Start 启用服务器
func (h *Server) Start(cnf conf.Conf) (err error) {
	h.serverName = cnf.String("name")
	h.serverType = cnf.String("type")
	h.logger = logger.New(fmt.Sprintf("%s-%s-%s", h.domain, h.serverName, h.serverType))
	h.logger.Infof("开始启动:%s.%s(%s)", h.serverName, h.serverType, h.tag)
	if strings.EqualFold(cnf.String("status"), servers.ST_STOP) {
		return fmt.Errorf("server未启动:%s(%s) 配置为:%s", cnf.String("name"), h.serverType, cnf.String("status"))
	}

	h.engines = cnf.String("engines")
	h.engineNames = cnf.Strings("engines", []string{})
	h.engineNames = append(h.engineNames, []string{"go", "rpc"}...)
	h.serviceRegistry, err = service.NewRegister(h.runMode, h.domain, h.serverName, h.logger, h.crurrentRegistryAddress, h.crossRegistryAddress)
	if err != nil {
		return fmt.Errorf("register初始化失败 mode:%s,domain:%s(err:%v)", h.serverType, h.domain, err)
	}

	// 启动执行引擎
	h.engine, err = engines.NewServiceEngine(h.domain, h.serverName, h.serverType, h.registry, h.logger, h.engineNames...)
	if err != nil {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(%s)(err:%v)", h.domain, h.serverName, h.serverType, err)
	}

	if !servers.IsDebug && strings.EqualFold(h.serverType, servers.SRV_TP_RPC) && len(h.engine.GetServices()) == 0 {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(%s)(err:engine中未找到任何服务)", h.domain, h.serverName, h.serverType)
	}

	//构建服务器
	h.server, err = servers.NewRegistryServer(h.serverType, h.engine, cnf, h.logger)
	if err != nil {
		return fmt.Errorf("server启动失败:%s.%s(%s)(err:%v)", h.serverName, h.serverType, h.tag, err)
	}
	err = h.server.Start()
	if err != nil {
		return fmt.Errorf("server启动失败:%s.%s(%s)(err:%v)", h.serverName, h.serverType, h.tag, err)
	}
	h.address = h.server.GetAddress()
	h.runTime = time.Now()
	h.localServices = h.server.GetServices()
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
	return p != h.engines
}

//Shutdown 关闭服务器
func (h *Server) Shutdown() {
	h.server.Shutdown()
	h.engine.Close()
}
