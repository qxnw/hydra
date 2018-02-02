package core

import (
	"strings"
)

//SetJWT Server
func (s *CoreServer) SetJWT(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	s.conf.JWTAuth.Enable = enable
	s.conf.JWTAuth.Name = name
	s.conf.JWTAuth.Secret = secret
	s.conf.JWTAuth.Mode = mode
	s.conf.JWTAuth.Exclude = exclude
	s.conf.JWTAuth.ExpireAt = expireAt
}

//SetHost 设置组件的host name
func (s *CoreServer) SetHost(host string) {
	if len(host) == 0 {
		s.conf.Hosts = make([]string, 0, 0)
		return
	}
	s.conf.Hosts = strings.Split(host, ",")
	return
}

//SetMetric 重置metric
func (s *CoreServer) SetMetric(host string, dataBase string, userName string, password string, cron string) {
	err := s.Metric.Restart(host, dataBase, userName, password, cron, s.Logger)
	if err != nil {
		s.Errorf("%s启动metric失败：%v", s.conf.GetFullName(), err)
	}
}

//SetHeader 设置http头
func (s *CoreServer) SetHeader(headers map[string]string) {
	s.conf.Headers = headers
}

//StopMetric stop metric
func (s *CoreServer) StopMetric() {
	s.Metric.Stop()
}
