package http

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
	"github.com/qxnw/lib4go/logger"
)

//WebResponsiveServer web 响应式服务器
type WebResponsiveServer struct {
	*ApiResponsiveServer
	webServer *WebServer
}

//NewWebResponsiveServer 构建基于注册中心的响应式web服务器
func NewWebResponsiveServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *WebResponsiveServer, err error) {
	h = &WebResponsiveServer{
		ApiResponsiveServer: &ApiResponsiveServer{},
	}
	h.engine = engine
	h.closeChan = make(chan struct{})
	h.currentConf = responsive.NewResponsiveConfBy(xconf.NewJSONConfWithEmpty(), cnf)
	h.Logger = logger
	h.pubs = make([]string, 0, 2)
	if h.webServer, err = NewWebServer(h.currentConf.ServerConf, nil, WithIP(h.currentConf.IP), WithLogger(logger)); err != nil {
		return
	}
	h.server = h.webServer
	if err = h.SetConf(true, h.currentConf); err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *WebResponsiveServer) Restart(cnf *responsive.ResponsiveConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.closeChan = make(chan struct{})
	w.done = false
	w.once = sync.Once{}
	if w.server, err = NewWebServer(w.currentConf.ServerConf, nil, WithIP(w.currentConf.IP), WithLogger(w.Logger)); err != nil {
		return
	}

	if err = w.SetConf(true, cnf); err != nil {
		return
	}
	return w.Start()
}
