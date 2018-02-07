package standard

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetMetric 重置metric
func (s *Server) SetMetric(host string, dataBase string, userName string, password string, cron string) {
	err := s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
	if err != nil {
		s.Errorf("%s启动metric失败：%v", s.conf.GetFullName(), err)
	}
}

//StopMetric stop metric
func (s *Server) StopMetric() {
	s.metric.Stop()
}

//SetQueues 设置监听队列
func (s *Server) SetQueues(raw string, queues ...*conf.Queue) (err error) {
	s.Processor, err = s.getProcessor(s.conf.Get("proto"), raw, queues)
	return err
}
