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
func QueueMessageCountCollect(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		if err = ctx.Request.Setting.Check("queue", "key"); err != nil {
			response.SetStatus(500)
			return
		}
		title := ctx.Request.Setting.GetString("title", "消息队列监控")
		msg := ctx.Request.Setting.GetString("msg", "消息队列服务:@key当前值:@current")
		platform := ctx.Request.Setting.GetString("platform", "----")
		queueName := ctx.Request.Setting.GetString("queue")
		key := ctx.Request.Setting.GetString("key")
		max := ctx.Request.Setting.GetInt64("max")
		min := ctx.Request.Setting.GetInt64("min")

		queue, err := c.GetQueue(queueName)
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
		tf.Set("key", ctx.Request.Setting.GetString("key"))
		tf.Set("value", strconv.Itoa(result))
		tf.Set("current", fmt.Sprintf("%d", value))
		tf.Set("level", ctx.Request.Setting.GetString("level", "1"))
		tf.Set("group", ctx.Request.Setting.GetString("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("{@key}"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, result, "queue-count")
		response.SetContent(st, err)
		return
	}
}
