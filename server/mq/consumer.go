package mq

import (
	"os"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/mq"
)

type taskOption struct {
	metric *InfluxMetric
	ip     string
	logger context.Logger
}

//TaskOption 任务设置选项
type TaskOption func(*taskOption)

//WithLogger 设置日志记录组件
func WithLogger(logger context.Logger) TaskOption {
	return func(o *taskOption) {
		o.logger = logger
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
	registry   context.IServiceRegistry
	handlers   []Handler
	serverName string
	*taskOption
}

//NewMQConsumer 构建服务器
func NewMQConsumer(name string, address string, version string, opts ...TaskOption) (s *MQConsumer, err error) {
	s = &MQConsumer{serverName: name}
	s.taskOption = &taskOption{}
	for _, opt := range opts {
		opt(s.taskOption)
	}
	if s.taskOption.logger == nil {
		s.taskOption.logger = NewLogger(name, os.Stdout)
	}
	s.taskOption.metric = NewInfluxMetric()
	s.handlers = append(s.handlers, s.metric)
	s.handlers = append(s.handlers, []Handler{Logging(), Recovery()}...)
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
		r := &Context{msg: m, server: s, params: m.GetMessage(), handle: handle}
		r.invoke()
		if r.err == nil {
			m.Ack()
		}
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
