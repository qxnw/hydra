package cron

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
	"github.com/qxnw/lib4go/logger"
)

//CronResponsiveServer rpc 响应式服务器
type CronResponsiveServer struct {
	server        *CronServer
	engine        servers.IRegistryEngine
	pubs          []string
	shardingIndex int
	shardingCount int
	master        bool
	currentConf   *responsive.ResponsiveConf
	closeChan     chan struct{}
	once          sync.Once
	done          bool
	pubLock       sync.Mutex
	*logger.Logger
	mu sync.Mutex
}

//NewCronResponsiveServer 创建rpc服务器
func NewCronResponsiveServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *CronResponsiveServer, err error) {
	h = &CronResponsiveServer{engine: engine,
		closeChan:   make(chan struct{}),
		currentConf: responsive.NewResponsiveConfBy(xconf.NewJSONConfWithEmpty(), cnf),
		Logger:      logger,
		pubs:        make([]string, 0, 2),
	}
	h.server, err = NewCronServer(h.currentConf.ServerConf, "", nil, WithIP(h.currentConf.IP), WithLogger(logger))
	if err != nil {
		return
	}
	err = h.SetConf(true, h.currentConf)
	if err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *CronResponsiveServer) Restart(cnf *responsive.ResponsiveConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.closeChan = make(chan struct{})
	w.done = false
	w.currentConf = cnf
	w.once = sync.Once{}
	w.server, err = NewCronServer(w.currentConf.ServerConf, "", nil, WithIP(w.currentConf.IP), WithLogger(w.Logger))
	if err != nil {
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
func (w *CronResponsiveServer) Start() (err error) {
	err = w.server.Run()
	if err != nil {
		return
	}
	return w.publish()
}

//Shutdown 关闭服务器
func (w *CronResponsiveServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
	})
	w.unpublish()
	w.server.Shutdown(time.Second)
}

//GetAddress 获取服务器地址
func (w *CronResponsiveServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *CronResponsiveServer) GetStatus() string {
	return w.server.GetStatus()
}

//GetServices 获取服务列表
func (w *CronResponsiveServer) GetServices() []string {
	return w.engine.GetServices()
}
