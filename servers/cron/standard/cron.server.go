package standard

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
)

//Server cron服务器
type Server struct {
	*option
	conf *conf.ServerConf
	*Processor
	running string
	addr    string
}

//New 创建mqc服务器
func New(conf *conf.ServerConf, redisSetting string, tasks []*conf.Task, opts ...Option) (t *Server, err error) {
	t = &Server{conf: conf}
	t.option = &option{metric: middleware.NewMetric(t.conf)}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(conf.GetFullName(), logger.CreateSession())
	}
	if tasks != nil && len(tasks) > 0 {
		err = t.SetTasks(redisSetting, tasks)
	}
	return
}

// Run the http server
func (s *Server) Run() error {
	if s.running == servers.ST_RUNNING {
		return nil
	}
	s.running = servers.ST_RUNNING
	errChan := make(chan error, 1)
	go func(ch chan error) {
		if err := s.Processor.Start(); err != nil {
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
func (s *Server) Shutdown() {
	if s.Processor != nil {
		s.running = servers.ST_STOP
		s.Processor.Close()
		s.Warnf("%s:已关闭", s.conf.GetFullName())

	}
}

//Pause 暂停服务器
func (s *Server) Pause() {
	if s.Processor != nil {
		s.running = servers.ST_PAUSE
		s.Processor.Pause()
		time.Sleep(time.Second)
	}
}

//Resume 恢复执行
func (s *Server) Resume() error {
	if s.Processor != nil {
		s.running = servers.ST_RUNNING
		s.Processor.Resume()
	}
	return nil
}

//GetAddress 获取当前服务地址
func (s *Server) GetAddress() string {
	return fmt.Sprintf("cron://%s", s.conf.IP)
}

//GetStatus 获取当前服务器状态
func (s *Server) GetStatus() string {
	return s.running
}
