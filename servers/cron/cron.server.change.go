package cron

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetMetric 重置metric
func (s *CronServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	return s.metric.Restart(host, dataBase, userName, password, cron, s.Logger)
}

//StopMetric stop metric
func (s *CronServer) StopMetric() error {
	s.metric.Stop()
	return nil
}

//SetTasks 设置定时任务
func (s *CronServer) SetTasks(redisSetting string, tasks []*conf.Task) (err error) {
	s.Processor, err = s.getProcessor(redisSetting, tasks)
	return err
}
