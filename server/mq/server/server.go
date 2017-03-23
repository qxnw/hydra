package server

import "github.com/qxnw/lib4go/mq"

//StompConsumerServer 消息消费队列服务器
type StompConsumerServer struct {
	*mq.StompConsumer
}

//NewServer 构建服务器
func NewServer(address string, version string) (s *StompConsumerServer, err error) {
	s = &StompConsumerServer{}
	s.StompConsumer, err = mq.NewStompConsumer(mq.ConsumerConfig{Address: address, Version: version, Persistent: "persistent"})
	return
}

//Run 运行
func (s *StompConsumerServer) Run() error {
	return s.StompConsumer.Connect()
}

//Use 启用消息处理
func (s *StompConsumerServer) Use(queue string, handle func(mq.IMessage)) error {
	return s.StompConsumer.Consume(queue, handle)
}

//UnUse 启用消息处理
func (s *StompConsumerServer) UnUse(queue string) {
	s.StompConsumer.UnConsume(queue)
}

//Close 关闭服务器
func (s *StompConsumerServer) Close() {
	s.StompConsumer.Close()
}
