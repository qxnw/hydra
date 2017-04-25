package mq

import (
	"sync"
	"time"

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/utility"
)

type taskOption struct {
	metric       *InfluxMetric
	ip           string
	registryRoot string
	registry     server.IServiceRegistry
}

//TaskOption 任务设置选项
type TaskOption func(*taskOption)

//WithRegistry 添加服务注册组件
func WithRegistry(r server.IServiceRegistry, root string) TaskOption {
	return func(o *taskOption) {
		o.registry = r
		o.registryRoot = root
	}
}

//WithIP 设置为循环执行任务
func WithIP(ip string) TaskOption {
	return func(o *taskOption) {
		o.ip = ip
	}
}

//MQConsumer 消息消费队列服务器
type MQConsumer struct {
	consumer    *mq.StompConsumer
	handlers    []Handler
	serverName  string
	p           *sync.Pool
	clusterPath string
	*logger.Logger
	*taskOption
}

//NewMQConsumer 构建服务器
func NewMQConsumer(name string, address string, version string, opts ...TaskOption) (s *MQConsumer, err error) {
	s = &MQConsumer{serverName: name, handlers: make([]Handler, 0, 3),
		Logger: logger.GetSession("hydra.mq", utility.GetGUID()),
		p: &sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
	}
	s.taskOption = &taskOption{metric: NewInfluxMetric()}
	for _, opt := range opts {
		opt(s.taskOption)
	}

	s.handlers = append(s.handlers,
		Logging(),
		Recovery(),
		s.metric)
	s.consumer, err = mq.NewStompConsumer(mq.ConsumerConfig{Address: address, Version: version, Persistent: "persistent"})
	return
}

//Run 运行
func (s *MQConsumer) Run() error {
	err := s.registryServer()
	if err != nil {
		return err
	}
	s.Infof("start mq server(%s)", s.serverName)
	err = s.consumer.ConnectLoop()
	if err != nil {
		s.unregistryServer()
	}
	return err
}

//SetInfluxMetric 重置metric
func (s *MQConsumer) SetInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) {
	s.metric.RestartReport(host, dataBase, userName, password, timeSpan)
}

//SetName 设置组件的server name
func (s *MQConsumer) SetName(name string) {
	s.serverName = name
}

//StopInfluxMetric stop metric
func (s *MQConsumer) StopInfluxMetric() {
	s.metric.Stop()
}

//Use 启用消息处理
func (s *MQConsumer) Use(queue string, handle func(*Context) error) error {
	s.Infof("start consume(%s, queue:%s)", s.serverName, queue)
	err := s.consumer.Consume(queue, func(m mq.IMessage) {
		r := s.p.Get().(*Context)
		message := m.GetMessage()
		r.reset(queue, m, s, message, handle)
		r.Logger = logger.GetSession(queue, utility.GetGUID())
		r.invoke()
		if r.err == nil && r.statusCode == 200 {
			m.Ack()
		}
		r.Close()
		s.p.Put(r)
	})
	if err != nil {
		s.Errorf("server:%s(err:%v)", s.serverName, err)
	}
	return nil
}

//UnUse 启用消息处理
func (s *MQConsumer) UnUse(queue string) {
	s.consumer.UnConsume(queue)
}

//Close 关闭服务器
func (s *MQConsumer) Close() {
	s.Errorf("mq server(%s) closed", s.serverName)
	s.unregistryServer()
	s.consumer.Close()
}
