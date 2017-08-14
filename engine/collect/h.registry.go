package collect

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) registryCollect(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	title := ctx.Input.GetArgValue("title", "注册中心服务")
	msg := ctx.Input.GetArgValue("msg", "注册中心服务:@url当前数量:@current")
	path, err := ctx.Input.GetArgByName("path")
	if err != nil {
		return
	}
	minValue, err := ctx.Input.GetArgIntValue("min")
	if err != nil {
		return
	}
	data, _, err := s.registry.GetChildren(path)
	if err != nil {
		return
	}
	value := 1
	if len(data) >= minValue {
		value = 0
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("host", path)
	tf.Set("url", path)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", strconv.Itoa(len(data)))
	tf.Set("level", ctx.Input.GetArgValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "registry", tf, value)
	response.Set(st, err)
	return
}
