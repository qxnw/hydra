package component

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/queue"
)

//IComponentQueue Component Queue
type IComponentQueue interface {
	GetQueue(names ...string) (q queue.IQueue, err error)
	Close() error
}

//StandardQueue queue
type StandardQueue struct {
	IContainer
	name       string
	queueCache cmap.ConcurrentMap
}

//NewStandardQueue 创建queue
func NewStandardQueue(c IContainer, name ...string) *StandardQueue {
	if len(name) > 0 {
		return &StandardQueue{IContainer: c, name: name[0], queueCache: cmap.New(2)}
	}
	return &StandardQueue{IContainer: c, name: "queue", queueCache: cmap.New(2)}
}

//GetQueue GetQueue
func (s *StandardQueue) GetQueue(names ...string) (q queue.IQueue, err error) {
	name := s.name
	if len(names) > 0 {
		name = names[0]
	}
	queueConf, err := s.IContainer.GetVarConf("queue", name)
	if err != nil {
		return nil, fmt.Errorf("../var/queue/%s %v", name, err)
	}
	key := fmt.Sprintf("%s:%d", name, queueConf.GetVersion())

	_, iqueue, err := s.queueCache.SetIfAbsentCb(key, func(input ...interface{}) (d interface{}, err error) {
		queueConf := input[0].(*conf.JSONConf)
		var qConf conf.QueueConf
		if err = queueConf.Unmarshal(&qConf); err != nil {
			return nil, err
		}
		if b, err := govalidator.ValidateStruct(&qConf); !b {
			return nil, err
		}
		return queue.NewQueue(qConf.Address, string(queueConf.GetRaw()))
	}, queueConf)
	if err != nil {
		err = fmt.Errorf("创建queue失败:%s,err:%v", string(queueConf.GetRaw()), err)
		return
		return
	}
	q = iqueue.(queue.IQueue)
	return

}

//Close 释放所有缓存配置
func (s *StandardQueue) Close() error {
	s.queueCache.RemoveIterCb(func(k string, v interface{}) bool {
		v.(queue.IQueue).Close()
		return true
	})
	return nil
}
