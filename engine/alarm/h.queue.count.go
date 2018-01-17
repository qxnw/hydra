package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/queue"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) queueCountCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
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
	queue, err := getQueue(ctx, queueName)
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
	st, err := s.checkAndSave(ctx, "queue-count", tf, result)
	response.SetError(st, err)
	return
}

func getQueue(ctx *context.Context, name string) (q queue.IQueue, err error) {
	_, iqueue, err := queueCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		content, err := ctx.Input.GetVarParam("queue", name)
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
