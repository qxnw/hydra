package main

import (
	"fmt"

	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/engine"
	_ "github.com/qxnw/hydra/engine/script"

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
	serviceRegister service.IRegister
	registry        string
}

//NewHydraServer 初始化服务器
func NewHydraServer(domain string, runMode string, registry string) *HydraServer {
	return &HydraServer{
		domain:   domain,
		runMode:  runMode,
		registry: registry,
		engine:   engine.NewStandardEngine(),
	}
}

//Start 启用服务器
func (h *HydraServer) Start(cnf conf.Conf) (err error) {
	tp := cnf.String("type")
	name := cnf.String("name")

	//构建服务器
	h.server, err = server.NewServer(tp, h.engine, nil, cnf)
	if err != nil {
		return fmt.Errorf("server启动失败 type:%s name:%s(err:%v)", tp, name, err)
	}

	// 启动执行引擎
	svs, err := h.engine.Start(h.domain, name, tp)
	if err != nil {
		return fmt.Errorf("engine启动失败 domain:%s name:%s(err:%v)", h.domain, name, err)
	}
	err = h.server.Start()
	if err != nil {
		return err
	}

	//启动注册中心
	if h.runMode == mode_cluster {
		h.serviceRegister, err = service.NewRegister(tp, h.domain, name, strings.Split(h.registry, ",")...)
		if err != nil {
			return fmt.Errorf("register初始化失败 mode:%s,domain:%s(err:%v)", tp, h.domain, err)
		}
		for _, v := range svs {
			h.serviceRegister.Register(v, h.server.GetAddress(), h.server.GetAddress())
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
	h.server.Shutdown()
}
