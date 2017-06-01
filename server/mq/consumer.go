package mq

import (
	"sync"
	"time"

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/mq"
)

type taskOption struct {
	metric       *InfluxMetric
	ip           string
	registryRoot string
	registry     server.IServiceRegistry
	version      string
	*logger.Logger
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

//WithVersion 设置MQ版本号
func WithVersion(version string) TaskOption {
	return func(o *taskOption) {
		o.version = version
	}
}

//MQConsumer 消息消费队列服务器
type MQConsumer struct {
	address     string
	consumer    mq.MQConsumer
	handlers    []Handler
	serverName  string
	p           *sync.Pool
	clusterPath string
	running     bool
	*taskOption
}

//NewMQConsumer 构建服务器
func NewMQConsumer(name string, address string, opts ...TaskOption) (s *MQConsumer, err error) {
	s = &MQConsumer{serverName: name, address: address, handlers: make([]Handler, 0, 3),
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
	if s.taskOption.Logger == nil {
		s.taskOption.Logger = logger.GetSession("hydra.mq", logger.CreateSession())
	}
	s.handlers = append(s.handlers,
		Logging(),
		Recovery(),
		s.metric)
	s.consumer, err = mq.NewMQConsumer(address, mq.WithVersion(s.version))
	return
}

//Run 运行
func (s *MQConsumer) Run() error {
	err := s.consumer.Connect()
	if err != nil {
		s.running = false
		s.unregistryServer()
		return err
	}
	s.Infof("Connected to %s", s.address)
	s.running = true
	return s.registryServer()
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
	//s.Infof("start consume(%s/%s)", s.serverName, queue)
	err := s.consumer.Consume(queue, func(m mq.IMessage) {
		r := s.p.Get().(*Context)
		message := m.GetMessage()
		r.reset(queue, m, s, message, handle)
		r.Logger = logger.GetSession(queue, logger.CreateSession())
		r.invoke()
		if r.statusCode == 200 {
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
	s.running = false
	s.unregistryServer()
	s.consumer.Close()
	s.Infof("mq: Server closed(%s)", s.serverName)

}
