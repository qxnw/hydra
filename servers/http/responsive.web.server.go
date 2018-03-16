package http

import (
	"fmt"
	"sync"
	"time"

	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/logger"
)

//WebResponsiveServer web 响应式服务器
type WebResponsiveServer struct {
	*ApiResponsiveServer
	webServer *WebServer
}

//NewWebResponsiveServer 构建基于注册中心的响应式web服务器
func NewWebResponsiveServer(registryAddr string, cnf conf.IServerConf, logger *logger.Logger) (h *WebResponsiveServer, err error) {
	h = &WebResponsiveServer{
		ApiResponsiveServer: &ApiResponsiveServer{registryAddr: registryAddr},
	}
	h.closeChan = make(chan struct{})
	h.currentConf = cnf
	h.Logger = logger
	h.pubs = make([]string, 0, 2)
	// 启动执行引擎
	h.engine, err = engines.NewServiceEngine(cnf, registryAddr, h.Logger, cnf.GetStrings("engines", "go", "rpc")...)
	if err != nil {
		return nil, fmt.Errorf("%s:engine启动失败%v", cnf.GetServerName(), err)
	}

	if h.webServer, err = NewWebServer(cnf.GetServerName(), cnf.GetString("address", ":8080"), nil, WithLogger(logger)); err != nil {
		return
	}
	h.server = h.webServer
	if err = h.SetConf(true, h.currentConf); err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *WebResponsiveServer) Restart(cnf conf.IServerConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.closeChan = make(chan struct{})
	w.done = false
	w.once = sync.Once{}
	// 启动执行引擎
	w.engine, err = engines.NewServiceEngine(cnf, w.registryAddr, w.Logger, cnf.GetStrings("engines", "go", "rpc")...)
	if err != nil {
		return fmt.Errorf("%s:engine启动失败%v", cnf.GetServerName(), err)
	}

	if w.server, err = NewWebServer(cnf.GetServerName(), cnf.GetString("address", ":8080"), nil, WithLogger(w.Logger)); err != nil {
		return
	}

	if err = w.SetConf(true, cnf); err != nil {
		return
	}
	return w.Start()
}
