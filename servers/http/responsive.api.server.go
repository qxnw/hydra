package http

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/responsive"
	"github.com/qxnw/lib4go/logger"
)

type IServer interface {
	Run(address ...interface{}) error
	Shutdown(timeout time.Duration)
	GetStatus() string
	GetAddress() string
	SetRouters(routers []*conf.Router) (err error)
	SetJWT(auth *conf.Auth) error
	SetAjaxRequest(allow bool) error
	SetHosts(hosts []string) error
	SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) error
	SetMetric(host string, dataBase string, userName string, password string, cron string) error
	SetHeader(headers map[string]string) error
	StopMetric() error
}

//ApiResponsiveServer api 响应式服务器
type ApiResponsiveServer struct {
	server      IServer
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

//NewApiResponsiveServer 创建API服务器
func NewApiResponsiveServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *ApiResponsiveServer, err error) {
	h = &ApiResponsiveServer{engine: engine,
		closeChan:   make(chan struct{}),
		currentConf: responsive.NewResponsiveConfBy(xconf.NewJSONConfWithEmpty(), cnf),
		Logger:      logger,
		pubs:        make([]string, 0, 2),
	}
	h.server, err = NewApiServer(h.currentConf.ServerConf, nil, WithIP(h.currentConf.IP), WithLogger(logger))
	if err != nil {
		return
	}
	err = h.SetConf(h.currentConf)
	if err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *ApiResponsiveServer) Restart(cnf *responsive.ResponsiveConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.closeChan = make(chan struct{})
	w.currentConf = cnf
	w.once = sync.Once{}
	w.server, err = NewApiServer(w.currentConf.ServerConf, nil, WithIP(w.currentConf.IP), WithLogger(w.Logger))
	if err != nil {
		return
	}
	err = w.SetConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//Start 启用服务
func (w *ApiResponsiveServer) Start() (err error) {
	err = w.server.Run(w.currentConf.GetString("address", ":80"))
	if err != nil {
		return
	}
	return w.publish()
}

//Shutdown 关闭服务器
func (w *ApiResponsiveServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
	})
	w.unpublish()
	w.server.Shutdown(time.Duration(w.currentConf.GetInt("timeout", 10)) * time.Second)
}

//GetAddress 获取服务器地址
func (w *ApiResponsiveServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *ApiResponsiveServer) GetStatus() string {
	return w.server.GetStatus()
}

//GetServices 获取服务列表
func (w *ApiResponsiveServer) GetServices() []string {
	return w.engine.GetServices()
}
