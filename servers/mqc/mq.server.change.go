package mqc

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetMetric 重置metric
func (s *MqcServer) SetMetric(host string, dataBase string, userName string, password string, cron string) error {
	s.metric.Stop()
	if err := s.metric.Restart(host, dataBase, userName, password, cron, s.Logger); err != nil {
		err = fmt.Errorf("metric设置有误:%v", err)
		return err
	}
	return nil

}

//StopMetric stop metric
func (s *MqcServer) StopMetric() error {
	s.metric.Stop()
	return nil
}

//SetQueues 设置监听队列
func (s *MqcServer) SetQueues(raw string, queues []*conf.Queue) (err error) {
	s.Processor, err = s.getProcessor(s.conf.Get("proto"), raw, queues)
	if err != nil {
		err = fmt.Errorf("queue设置有误:%v", err)
		return err
	}
	return nil
}
