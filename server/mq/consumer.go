package mq

import "github.com/qxnw/lib4go/mq"

//MQConsumer 消息消费队列服务器
type MQConsumer struct {
	consumer *mq.StompConsumer
}

//NewMQConsumer 构建服务器
func NewMQConsumer(address string, version string) (s *MQConsumer, err error) {
	s = &MQConsumer{}
	s.consumer, err = mq.NewStompConsumer(mq.ConsumerConfig{Address: address, Version: version, Persistent: "persistent"})
	return
}

//Run 运行
func (s *MQConsumer) Run() error {
	return s.consumer.Connect()
}

//Use 启用消息处理
func (s *MQConsumer) Use(queue string, handle func(mq.IMessage)) error {
	return s.consumer.Consume(queue, handle)
}

//UnUse 启用消息处理
func (s *MQConsumer) UnUse(queue string) {
	s.consumer.UnConsume(queue)
}

//Close 关闭服务器
func (s *MQConsumer) Close() {
	s.consumer.Close()
}
