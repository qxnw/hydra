package mqc

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/responsive"
	"github.com/qxnw/lib4go/logger"
)

//MqcResponsiveServer rpc 响应式服务器
type MqcResponsiveServer struct {
	server        *MqcServer
	engine        servers.IRegistryEngine
	pubs          []string
	currentConf   *responsive.ResponsiveConf
	closeChan     chan struct{}
	once          sync.Once
	done          bool
	shardingIndex int
	shardingCount int
	master        bool
	pubLock       sync.Mutex
	*logger.Logger
	mu sync.Mutex
}

//NewMqcResponsiveServer 创建mqc服务器
func NewMqcResponsiveServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *MqcResponsiveServer, err error) {
	h = &MqcResponsiveServer{engine: engine,
		closeChan:   make(chan struct{}),
		currentConf: responsive.NewResponsiveConfBy(xconf.NewJSONConfWithEmpty(), cnf),
		Logger:      logger,
		pubs:        make([]string, 0, 2),
	}
	if h.server, err = NewMqcServer(h.currentConf.ServerConf, "", nil, WithIP(h.currentConf.IP), WithLogger(logger)); err != nil {
		return
	}
	if err = h.SetConf(true, h.currentConf); err != nil {
		return
	}
	return
}

//Restart 重启服务器
func (w *MqcResponsiveServer) Restart(cnf *responsive.ResponsiveConf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.done = false
	w.closeChan = make(chan struct{})
	w.once = sync.Once{}
	if w.server, err = NewMqcServer(w.currentConf.ServerConf, "", nil, WithIP(w.currentConf.IP), WithLogger(w.Logger)); err != nil {
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
func (w *MqcResponsiveServer) Start() (err error) {
	if err = w.server.Run(); err != nil {
		return
	}
	return w.publish()
}

//Shutdown 关闭服务器
func (w *MqcResponsiveServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
	})
	w.unpublish()
	timeout := w.currentConf.GetInt("timeout", 10)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

//GetAddress 获取服务器地址
func (w *MqcResponsiveServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *MqcResponsiveServer) GetStatus() string {
	return w.server.GetStatus()
}

//GetServices 获取服务列表
func (w *MqcResponsiveServer) GetServices() []string {
	return w.engine.GetServices()
}
