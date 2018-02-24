package rpc

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *RpcServer) SetRouters(routers []*conf.Router) (err error) {
	if s.Processor, err = s.getProcessor(routers); err != nil {
		return
	}
	return
}

//SetJWT Server
func (s *RpcServer) SetJWT(auth *conf.Auth) error {
	s.conf.SetMetadata("jwt", auth)
	return nil
}

//SetHosts 设置组件的host name
func (s *RpcServer) SetHosts(hosts []string) error {
	if len(hosts) == 0 {
		s.conf.Hosts = make([]string, 0, 0)
		return nil
	}
	s.conf.Hosts = hosts
	return nil
}

//SetMetric 重置metric
func (s *RpcServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	return s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
}

//StopMetric stop metric
func (s *RpcServer) StopMetric() error {
	s.metric.Stop()
	return nil
}

//SetHeader 设置http头
func (s *RpcServer) SetHeader(headers map[string]string) error {
	s.conf.Headers = headers
	return nil
}
