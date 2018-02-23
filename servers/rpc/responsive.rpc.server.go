package rpc

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
	"github.com/qxnw/lib4go/logger"
)

//RpcResponsiveServer rpc 响应式服务器
type RpcResponsiveServer struct {
	server      *RpcServer
	engine      servers.IRegistryEngine
	pubs        []string
	currentConf *responsive.ResponsiveConf
	closeChan   chan struct{}
	once        sync.Once
	done        bool
	pubLock     sync.Mutex
	*logger.Logger
	mu sync.Mutex
}

//NewRpcResponsiveServer 创建rpc服务器
func NewRpcResponsiveServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *RpcResponsiveServer, err error) {
	h = &RpcResponsiveServer{engine: engine,
		closeChan:   make(chan struct{}),
		currentConf: responsive.NewResponsiveConfBy(xconf.NewJSONConfWithEmpty(), cnf),
		Logger:      logger,
		pubs:        make([]string, 0, 2),
	}
	if h.server, err = NewRpcServer(h.currentConf.ServerConf, nil, WithIP(h.currentConf.IP), WithLogger(logger)); err != nil {
		return
	}
	if err = h.SetConf(true, h.currentConf); err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *RpcResponsiveServer) Restart(cnf *responsive.ResponsiveConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.done = false
	w.closeChan = make(chan struct{})
	w.currentConf = cnf
	w.once = sync.Once{}
	if w.server, err = NewRpcServer(w.currentConf.ServerConf, nil, WithIP(w.currentConf.IP), WithLogger(w.Logger)); err != nil {
		return
	}
	if err = w.SetConf(true, cnf); err != nil {
		return
	}
	if err = w.Start(); err == nil {
		w.currentConf = cnf
		return
	}
	return err
}

//Start 启用服务
func (w *RpcResponsiveServer) Start() (err error) {
	if err = w.server.Run(w.currentConf.GetString("address", ":81")); err != nil {
		return
	}
	return w.publish()
}

//Shutdown 关闭服务器
func (w *RpcResponsiveServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
	})
	w.unpublish()
	timeout := w.currentConf.GetInt("timeout", 10)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

//GetAddress 获取服务器地址
func (w *RpcResponsiveServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *RpcResponsiveServer) GetStatus() string {
	return w.server.GetStatus()
}

//GetServices 获取服务列表
func (w *RpcResponsiveServer) GetServices() []string {
	svs := w.engine.GetServices()
	nsevice := make([]string, 0, len(svs))
	for _, sv := range svs {
		if w.server.Find(sv) {
			nsevice = append(nsevice, sv)
		}
	}
	servers.Trace(w.Infof, w.currentConf.GetFullName(), "发布服务：", nsevice)
	return nsevice
}
