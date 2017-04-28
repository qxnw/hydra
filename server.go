package main

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/engine"
	_ "github.com/qxnw/hydra/engine/goplugin"
	_ "github.com/qxnw/hydra/engine/rpc_proxy"
	_ "github.com/qxnw/hydra/engine/script"
	"github.com/qxnw/lib4go/logger"

	"strings"

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/hydra/service"
)

var (
	mode_Standalone = "standalone"
	mode_cluster    = "cluster"
)

//HydraServer hydra server
type HydraServer struct {
	domain          string
	runMode         string
	engine          engine.IEngine
	server          server.IHydraServer
	serviceRegistry service.IServiceRegistry
	registryAddress []string
	registry        string
	services        []string
	logger          *logger.Logger
}

//NewHydraServer 初始化服务器
func NewHydraServer(domain string, runMode string, registry string, logger *logger.Logger, registryAddress []string) *HydraServer {
	return &HydraServer{
		domain:          domain,
		runMode:         runMode,
		registry:        registry,
		services:        make([]string, 0, 16),
		registryAddress: registryAddress,
		engine:          engine.NewStandardEngine(),
		logger:          logger,
	}
}

//Start 启用服务器
func (h *HydraServer) Start(cnf conf.Conf) (err error) {
	if strings.EqualFold(cnf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器:%s 配置为:%s", cnf.String("name"), cnf.String("status"))
	}
	serverName := cnf.String("name")
	tp := cnf.String("type")

	h.serviceRegistry, err = service.NewRegister(h.runMode, h.domain, serverName, h.logger, h.registryAddress)
	if err != nil {
		return fmt.Errorf("register初始化失败 mode:%s,domain:%s(err:%v)", tp, h.domain, err)
	}
	// 启动服务引擎
	svs, err := h.engine.Start(h.domain, serverName, tp, h.registry)
	if err != nil {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(err:%v)", h.domain, serverName, err)
	}
	if len(svs) == 0 {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(err:未找到服务)", h.domain, serverName)
	}
	//构建服务器
	h.server, err = server.NewServer(tp, h.engine, h.serviceRegistry, cnf)
	if err != nil {
		return fmt.Errorf("server启动失败:%s(err:%v)", serverName, err)
	}
	err = h.server.Start()
	if err != nil {
		return err
	}
	//注册服务列表
	if strings.EqualFold(tp, server.SRV_TP_RPC) {
		for _, v := range svs {
			path, err := h.serviceRegistry.Register(v, strings.Replace(h.server.GetAddress(), "//", "", -1), h.server.GetAddress())
			if err != nil {
				return err
			}
			h.services = append(h.services, path)
		}
	}
	return nil
}

//Notify 配置发生变化通知服务器变更
func (h *HydraServer) Notify(cnf conf.Conf) error {
	return h.server.Notify(cnf)
}

//GetStatus 获取当前服务状态
func (h *HydraServer) GetStatus() string {
	return h.server.GetStatus()
}

//Shutdown 关闭服务器
func (h *HydraServer) Shutdown() {
	if h.serviceRegistry != nil {
		for _, v := range h.services {
			h.serviceRegistry.Unregister(v)
		}
	}
	h.engine.Close()
	h.server.Shutdown()
}
