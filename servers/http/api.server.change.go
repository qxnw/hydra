package http

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *ApiServer) SetRouters(routers []*conf.Router) (err error) {
	s.engine.Handler, err = s.getHandler(routers)
	return
}

//SetJWT Server
func (s *ApiServer) SetJWT(auth *conf.Auth) error {
	s.conf.SetMeta("jwt", auth)
	return nil
}

//SetAjaxRequest 只允许ajax请求
func (s *ApiServer) SetAjaxRequest(allow bool) error {
	s.conf.SetMeta("ajax-request", allow)
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
	s.static.FilterExts = exts
	s.static.Prepare()
	return nil
}

//SetMetric 重置metric
func (s *ApiServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	return s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
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
