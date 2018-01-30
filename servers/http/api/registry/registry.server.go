package registry

import (
	"sync"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http"
	"github.com/qxnw/hydra/servers/http/api/standard"
	"github.com/qxnw/lib4go/logger"
)

//RegistryServer api 服务器
type RegistryServer struct {
	server *standard.Server
	engine servers.IRegistryEngine
	conf   conf.Conf
	*logger.Logger
	mu sync.Mutex
}

//NewRegistryServer 创建API服务器
func NewRegistryServer(engine servers.IRegistryEngine, cnf conf.Conf, logger *logger.Logger) (h *RegistryServer, err error) {
	serverConf := http.NewServerConf(cnf)
	h = &RegistryServer{engine: engine,
		conf:   conf.NewJSONConfWithEmpty(),
		Logger: logger,
		server: standard.New(serverConf, nil, standard.WithIP(serverConf.IP), standard.WithLogger(logger))}
	err = h.SetConf(NewRegistryConf(conf.NewJSONConfWithEmpty(), cnf))
	if err != nil {
		return
	}
	h.conf = cnf
	return
}

//Restart 重启服务器
func (w *RegistryServer) Restart(cnf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	serverConf := http.NewServerConf(cnf)
	w.conf = conf.NewJSONConfWithEmpty()
	w.server = standard.New(serverConf, nil, standard.WithIP(serverConf.IP), standard.WithLogger(w.Logger))
	err = w.SetConf(NewRegistryConf(conf.NewJSONConfWithEmpty(), cnf))
	if err != nil {
		return
	}
	w.conf = cnf
	return w.Start()
}

//Start 启用服务
func (w *RegistryServer) Start() (err error) {
	startChan := make(chan error, 1)
	tls, err := w.conf.GetSection("tls")
	if err != nil {
		go func(ch chan error) {
			err = w.server.Run(w.conf.String("address", ":81"))
			startChan <- err
		}(startChan)
	} else {
		go func(tls conf.Conf, ch chan error) {
			err = w.server.RunTLS(tls.String("cert"), tls.String("key"), tls.String("address", ":9898"))
			startChan <- err
		}(tls, startChan)
	}
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-startChan:
		return err
	}
}

//Shutdown 关闭服务器
func (w *RegistryServer) Shutdown() {
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
