package mq

import (
	"sync"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/utility"
)

type taskOption struct {
	metric   *InfluxMetric
	ip       string
	registry context.IServiceRegistry
}

//TaskOption 任务设置选项
type TaskOption func(*taskOption)

/*
//WithLogger 设置日志记录组件
func WithLogger(logger logger.ILogger) TaskOption {
	return func(o *taskOption) {
		o.logger = logger
	}
}
*/
//WithRegistry 设置服务注册组件
func WithRegistry(i context.IServiceRegistry) TaskOption {
	return func(o *taskOption) {
		o.registry = i
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
	consumer   *mq.StompConsumer
	handlers   []Handler
	serverName string
	p          *sync.Pool
	*taskOption
}

//NewMQConsumer 构建服务器
func NewMQConsumer(name string, address string, version string, opts ...TaskOption) (s *MQConsumer, err error) {
	s = &MQConsumer{serverName: name, handlers: make([]Handler, 3),
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
	return s.consumer.Connect()
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
	return s.consumer.Consume(queue, func(m mq.IMessage) {
		r := s.p.Get().(*Context)
		r.reset(m, s, m.GetMessage(), handle)
		r.Logger = logger.GetSession(queue, utility.GetGUID())
		r.invoke()
		if r.err == nil && r.statusCode == 200 {
			m.Ack()
		}
		r.Close()
		s.p.Put(r)
	})
}

//UnUse 启用消息处理
func (s *MQConsumer) UnUse(queue string) {
	s.consumer.UnConsume(queue)
}

//Close 关闭服务器
func (s *MQConsumer) Close() {
	s.consumer.Close()
}
