package http

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/circuit"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *ApiServer) SetRouters(routers []*conf.Router) (err error) {
	s.engine.Handler, err = s.getHandler(routers)
	return
}

//SetJWT Server
func (s *ApiServer) SetJWT(auth *conf.Auth) error {
	s.conf.SetMetadata("jwt", auth)
	return nil
}

//SetAjaxRequest 只允许ajax请求
func (s *ApiServer) SetAjaxRequest(allow bool) error {
	s.conf.SetMetadata("ajax-request", allow)
	return nil
}

//SetHosts 设置组件的host name
func (s *ApiServer) SetHosts(hosts []string) error {
	if len(hosts) == 0 {
		s.conf.Hosts = make([]string, 0, 0)
		return nil
	}
	s.conf.Hosts = hosts
	return nil
}

//SetStatic 设置静态文件路由
func (s *ApiServer) SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) error {
	s.static.Enable = enable
	s.static.Prefix = prefix
	s.static.RootPath = dir
	s.static.Exts = exts
	return nil
}

//SetMetric 重置metric
func (s *ApiServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	s.metric.Stop()
	if err := s.metric.Restart(host, dataBase, userName, password, cron, s.Logger); err != nil {
		err = fmt.Errorf("metric设置有误:%v", err)
		return err
	}
	return nil
}

//SetHeader 设置http头
func (s *ApiServer) SetHeader(headers map[string]string) error {
	s.conf.Headers = headers
	return nil
}

//StopMetric stop metric
func (s *ApiServer) StopMetric() error {
	s.metric.Stop()
	return nil
}

//CloseCircuitBreaker 关闭熔断配置
func (s *ApiServer) CloseCircuitBreaker() error {
	if c, ok := s.conf.GetMetadata("__circuit-breaker_").(*circuit.NamedCircuitBreakers); ok {
		c.Close()
	}
	return nil
}

//SetCircuitBreaker 设置熔断配置
func (s *ApiServer) SetCircuitBreaker(c *conf.CircuitBreaker) error {
	s.conf.SetMetadata("__circuit-breaker_", circuit.NewNamedCircuitBreakers(c))
	return nil
}
