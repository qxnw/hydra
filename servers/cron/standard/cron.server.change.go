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

//SetTasks 设置定时任务
func (s *Server) SetTasks(redisSetting string, tasks []*conf.Task) (err error) {
	s.Processor, err = s.getProcessor(redisSetting, tasks)
	return err
}
