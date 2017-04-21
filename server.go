package main

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/engine"
	_ "github.com/qxnw/hydra/engine/script"

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
	server          server.IHydraServer
	engine          engine.IEngine
	serviceRegistry service.IServiceRegistry
	registryAddress []string
	registry        string
	services        []string
}

//NewHydraServer 初始化服务器
func NewHydraServer(domain string, runMode string, registry string, registryAddress []string) *HydraServer {
	return &HydraServer{
		domain:          domain,
		runMode:         runMode,
		registry:        registry,
		services:        make([]string, 0, 16),
		registryAddress: registryAddress,
		engine:          engine.NewStandardEngine(),
	}
}

//Start 启用服务器
func (h *HydraServer) Start(cnf conf.Conf) (err error) {
	tp := cnf.String("type")
	serverName := cnf.String("name")
	if h.runMode != mode_Standalone {
		h.serviceRegistry, err = service.NewRegister(h.runMode, h.domain, serverName, h.registryAddress...)
		if err != nil {
			return fmt.Errorf("register初始化失败 mode:%s,domain:%s(err:%v)", tp, h.domain, err)
		}
	}
	//构建服务器
	h.server, err = server.NewServer(tp, h.engine, h.serviceRegistry, cnf)
	if err != nil {
		return fmt.Errorf("server启动失败:%s(err:%v)", serverName, err)
	}

	// 启动执行引擎
	svs, err := h.engine.Start(h.domain, serverName, tp)
	if err != nil {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(err:%v)", h.domain, serverName, err)
	}
	err = h.server.Start()
	if err != nil {
		return err
	}
	//启动注册中心
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

//Shutdown 关闭服务器
func (h *HydraServer) Shutdown() {
	if h.serviceRegistry != nil {
		for _, v := range h.services {
			h.serviceRegistry.Unregister(v)
		}
		h.serviceRegistry.Close()
	}
	h.server.Shutdown()
}
