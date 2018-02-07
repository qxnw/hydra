package component

import (
	"fmt"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/queue"
)

//IComponentQueue Component Queue
type IComponentQueue interface {
	GetDefaultQueue() (c queue.IQueue, err error)
	GetQueue(name string) (q queue.IQueue, err error)
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

//GetDefaultQueue 获取默然Queue
func (s *StandardQueue) GetDefaultQueue() (c queue.IQueue, err error) {
	return s.GetQueue(s.name)
}

//GetQueue GetQueue
func (s *StandardQueue) GetQueue(name string) (q queue.IQueue, err error) {
	_, iqueue, err := s.queueCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		content, err := s.IContainer.GetVarParam("queue", name)
		if err != nil {
			return nil, err
		}
		configMap, err := jsons.Unmarshal([]byte(content))
		if err != nil {
			return nil, err
		}
		address, ok := configMap["address"]
		if !ok {
			return nil, fmt.Errorf("queue配置文件错误，未包含address节点:var/queue/%s", name)
		}
		d, err = queue.NewQueue(address.(string), content)
		if err != nil {
			err = fmt.Errorf("创建queue失败:%s,err:%v", content, err)
			fmt.Println("queue.err:", err)
			return
		}

		return
	}, name)
	if err != nil {
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
