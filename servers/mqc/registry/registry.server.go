package registry

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/mqc/standard"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/logger"
)

//RegistryServer mqc 服务器
type RegistryServer struct {
	server        *standard.Server
	engine        servers.IRegistryEngine
	pubs          []string
	conf          xconf.Conf
	serverConf    *conf.ServerConf
	master        bool
	closeChan     chan struct{}
	shardingIndex int
	shardingCount int
	once          sync.Once
	done          bool
	pubLock       sync.Mutex
	*logger.Logger
	mu sync.Mutex
}

//NewRegistryServer 创建API服务器
func NewRegistryServer(engine servers.IRegistryEngine, cnf xconf.Conf, logger *logger.Logger) (h *RegistryServer, err error) {
	serverConf := conf.NewConfBy(cnf)
	h = &RegistryServer{engine: engine,
		closeChan:  make(chan struct{}),
		conf:       xconf.NewJSONConfWithEmpty(),
		serverConf: serverConf,
		Logger:     logger,
		pubs:       make([]string, 0, 2)}
	h.server, err = standard.New(serverConf, "", nil, standard.WithIP(serverConf.IP), standard.WithLogger(logger))
	if err != nil {
		return
	}
	err = h.SetConf(conf.NewRegistryConf(xconf.NewJSONConfWithEmpty(), cnf))
	if err != nil {
		return
	}
	h.conf = cnf
	return
}

//Restart 重启服务器
func (w *RegistryServer) Restart(cnf xconf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.closeChan = make(chan struct{})
	w.conf = xconf.NewJSONConfWithEmpty()
	w.once = sync.Once{}
	w.serverConf = conf.NewConfBy(cnf)
	w.server, err = standard.New(w.serverConf, "", nil, standard.WithIP(w.serverConf.IP), standard.WithLogger(w.Logger))
	if err != nil {
		return
	}
	err = w.SetConf(conf.NewRegistryConf(xconf.NewJSONConfWithEmpty(), cnf))
	if err != nil {
		return
	}
	w.conf = cnf
	return w.Start()
}

//Start 启用服务
func (w *RegistryServer) Start() (err error) {
	return w.publish()
}

//Shutdown 关闭服务器
func (w *RegistryServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
	})
	w.unpublish()
	timeout, _ := w.conf.Int("timeout", 10)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

//GetAddress 获取服务器地址
func (w *RegistryServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *RegistryServer) GetStatus() string {
	return w.server.GetStatus()
}

//GetServices 获取服务列表
func (w *RegistryServer) GetServices() []string {
	return w.engine.GetServices()
}
