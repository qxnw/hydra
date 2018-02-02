package registry

import (
	"sync"
	"time"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/logger"
)

//RegistryServer rpc 服务器
type RegistryServer struct {
	server     servers.IServer
	Engine     servers.IRegistryEngine
	serverConf *conf.ApiServerConf
	pubs       []string
	Conf       xconf.Conf
	closeChan  chan struct{}
	once       sync.Once
	done       bool
	pubLock    sync.Mutex
	*logger.Logger
	mu sync.Mutex
}

//NewRegistryServer 创建API服务器
func NewRegistryServer(engine servers.IRegistryEngine, serverConf *conf.ApiServerConf, server servers.IServer, cnf xconf.Conf, logger *logger.Logger) (h *RegistryServer) {

	h = &RegistryServer{Engine: engine,
		closeChan:  make(chan struct{}),
		Conf:       xconf.NewJSONConfWithEmpty(),
		serverConf: serverConf,
		Logger:     logger,
		pubs:       make([]string, 0, 2)}
	h.server = server
	return
}

//Restart 重启服务器
func (w *RegistryServer) Restart(cnf xconf.Conf) (err error) {
	return nil
}

//Start 启用服务
func (w *RegistryServer) Start() (err error) {
	err = w.server.Run(w.Conf.String("address", ":81"))
	if err != nil {
		return
	}
	return w.publish()
}

//Shutdown 关闭服务器
func (w *RegistryServer) Shutdown() {
	w.done = true
	w.once.Do(func() {
		close(w.closeChan)
		w.unpublish()
		timeout, _ := w.Conf.Int("timeout", 10)
		w.server.Shutdown(time.Duration(timeout) * time.Second)
	})
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
	return w.Engine.GetServices()
}
