package http

import (
	"github.com/qxnw/hydra/servers/pkg/circuit"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *WebServer) SetRouters(routers []*conf.Router) (err error) {
	s.engine.Handler, err = s.getHandler(routers)
	return
}

//SetJWT Server
func (s *WebServer) SetJWT(auth *conf.Auth) error {
	s.conf.SetMetadata("jwt", auth)
	return nil
}

//SetAjaxRequest 只允许ajax请求
func (s *WebServer) SetAjaxRequest(allow bool) error {
	s.conf.SetMetadata("ajax-request", allow)
	return nil
}

//SetHosts 设置组件的host name
func (s *WebServer) SetHosts(hosts []string) error {
	if len(hosts) == 0 {
		s.conf.Hosts = make([]string, 0, 0)
		return nil
	}
	s.conf.Hosts = hosts
	return nil
}

//SetStatic 设置静态文件路由
func (s *WebServer) SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) error {
	s.static.Enable = enable
	s.static.Prefix = prefix
	s.static.RootPath = dir
	s.static.Exts = exts
	return nil
}

//SetMetric 重置metric
func (s *WebServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	return s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
}

//SetHeader 设置http头
func (s *WebServer) SetHeader(headers map[string]string) error {
	s.conf.Headers = headers
	return nil
}

//StopMetric stop metric
func (s *WebServer) StopMetric() error {
	s.metric.Stop()
	return nil
}

//SetView 设置view参数
func (s *WebServer) SetView(view *conf.View) error {
	s.conf.SetMetadata("view", view)
	return nil
}

//CloseCircuitBreaker 关闭熔断配置
func (s *WebServer) CloseCircuitBreaker() error {
	if c, ok := s.conf.GetMetadata("__circuit-breaker_").(*circuit.NamedCircuitBreakers); ok {
		c.Close()
	}
	return nil
}

//SetCircuitBreaker 设置熔断配置
func (s *WebServer) SetCircuitBreaker(c *conf.CircuitBreaker) error {
	s.conf.SetMetadata("__circuit-breaker_", circuit.NewNamedCircuitBreakers(c))
	return nil
}
