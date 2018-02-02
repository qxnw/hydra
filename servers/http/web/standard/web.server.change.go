package standard

import (
	"strings"

	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *Server) SetRouters(routers []*conf.Router) (err error) {
	s.engine.Handler, err = s.getHandler(routers)
	return
}

//SetJWT Server
func (s *Server) SetJWT(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	s.conf.JWTAuth.Enable = enable
	s.conf.JWTAuth.Name = name
	s.conf.JWTAuth.Secret = secret
	s.conf.JWTAuth.Mode = mode
	s.conf.JWTAuth.Exclude = exclude
	s.conf.JWTAuth.ExpireAt = expireAt
}

//SetHost 设置组件的host name
func (s *Server) SetHost(host string) {
	if len(host) == 0 {
		s.conf.Hosts = make([]string, 0, 0)
		return
	}
	s.conf.Hosts = strings.Split(host, ",")
	return
}

//SetStatic 设置静态文件路由
func (s *Server) SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) {
	s.static.Enable = enable
	s.static.Prefix = prefix
	s.static.RootPath = dir
	s.static.FilterExts = exts
	s.static.Prepare()
}

//SetMetric 重置metric
func (s *Server) SetMetric(host string, dataBase string, userName string, password string, cron string) {
	err := s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
	if err != nil {
		s.Errorf("%s启动metric失败：%v", s.conf.GetFullName(), err)
	}
}

//SetHeader 设置http头
func (s *Server) SetHeader(headers map[string]string) {
	s.conf.Headers = headers
}

//SetView 设置view参数
func (s *Server) SetView(path, left, right string) {
	s.conf.View.Path = path
	s.conf.View.Left = left
	s.conf.View.Right = right
}

//StopMetric stop metric
func (s *Server) StopMetric() {
	s.metric.Stop()
}
