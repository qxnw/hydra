package mqc

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
)

//MqcServer mqc服务器
type MqcServer struct {
	*option
	conf *conf.ServerConf
	*Processor
	running string
	addr    string
}

//NewMqcServer 创建mqc服务器
func NewMqcServer(conf *conf.ServerConf, serverRaw string, queues []*conf.Queue, opts ...Option) (t *MqcServer, err error) {
	t = &MqcServer{conf: conf}
	t.option = &option{metric: middleware.NewMetric(t.conf)}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(conf.GetFullName(), logger.CreateSession())
	}
	if queues != nil && len(queues) > 0 {
		err = t.SetQueues(serverRaw, queues)
	}
	return
}

// Run the http server
func (s *MqcServer) Run() error {
	if s.running == servers.ST_RUNNING {
		return nil
	}
	s.running = servers.ST_RUNNING
	errChan := make(chan error, 1)
	if err := s.Processor.Consumes(); err != nil {
		return err
	}
	go func(ch chan error) {
		if err := s.Processor.Connect(); err != nil {
			ch <- err
		}
	}(errChan)
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		s.running = servers.ST_STOP
		return err
	}
}

//Shutdown 关闭服务器
func (s *MqcServer) Shutdown(timeout time.Duration) {
	if s.Processor != nil {
		s.running = servers.ST_STOP
		s.Processor.Close()
		time.Sleep(time.Second)
		s.Warnf("%s:已关闭", s.conf.GetFullName())

	}
}

//Pause 暂停服务器
func (s *MqcServer) Pause(timeout time.Duration) {
	if s.Processor != nil {
		s.running = servers.ST_PAUSE
		s.Processor.Close()
		time.Sleep(time.Second)
	}
}

//GetAddress 获取当前服务地址
func (s *MqcServer) GetAddress() string {
	return fmt.Sprintf("mqc://%s", s.conf.IP)
}

//GetStatus 获取当前服务器状态
func (s *MqcServer) GetStatus() string {
	return s.running
}
