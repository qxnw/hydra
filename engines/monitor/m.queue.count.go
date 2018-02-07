package monitor

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
)

//CollectQueueMessageCount 收集消息队列条数
func CollectQueueMessageCount(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		queueName, err := ctx.Request.Setting.Get("queue")
		if err != nil {
			return
		}
		key, err := ctx.Request.Setting.Get("key")
		if err != nil {
			return
		}
		queue, err := c.GetQueue(queueName)
		if err != nil {
			return
		}
		count, err := queue.Count(key)
		if err != nil {
			return
		}
		ip := xnet.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateredisListCount(c, ctx, count, "server", ip, "key", key)
		response.SetContent(0, err)
		return
	}
}
