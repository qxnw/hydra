package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

//QueueMessageCountCollect 队列里的消息条数收集
func QueueMessageCountCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		title := ctx.Input.GetArgsValue("title", "消息队列监控")
		msg := ctx.Input.GetArgsValue("msg", "消息队列服务:@key当前值:@current")
		platform := ctx.Input.GetArgsValue("platform", "----")
		queueName, err := ctx.Input.GetArgsByName("queue")
		if err != nil {
			return
		}
		key, err := ctx.Input.GetArgsByName("key")
		if err != nil {
			return
		}
		max := ctx.Input.GetArgsInt64("max")
		min := ctx.Input.GetArgsInt64("min")

		queue, err := component.NewStandardQueue(c, queueName).GetDefaultQueue()
		if err != nil {
			return
		}
		value, err := queue.Count(key)
		if err != nil {
			return
		}
		result := 1 //需要报警
		if ((min > 0 && value >= min) || min == 0) && ((max > 0 && value < max) || max == 0) {
			result = 0 //恢复
		}

		tf := transform.NewMap(map[string]string{})
		tf.Set("key", ctx.Input.GetArgsValue("key"))
		tf.Set("value", strconv.Itoa(result))
		tf.Set("current", fmt.Sprintf("%d", value))
		tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
		tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("{@key}"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, result, "queue-count")
		response.SetError(st, err)
		return
	}
}
