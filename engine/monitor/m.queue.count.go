package monitor

import (
	"energy/coupon-services/access"
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
	xnet "github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/queue"
)

func (s *monitorProxy) queueCountCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	redis, err := ctx.Input.GetArgsByName("queue")
	if err != nil {
		return
	}
	key, err := ctx.Input.GetArgsByName("key")
	if err != nil {
		return
	}
	queue, err := getQueue(redis)
	if err != nil {
		return
	}
	count, err := queue.Count(key)
	if err != nil {
		return
	}
	ip := xnet.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateredisListCount(ctx, count, "server", ip, "key", key)
	response.SetError(0, err)
	return
}

func getQueue(name string) (q queue.IQueue, err error) {
	_, iqueue, err := queueCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		content, err := access.Container.GetVarParam("queue", name)
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

var queueCache cmap.ConcurrentMap

func init() {
	queueCache = cmap.New(2)
}
