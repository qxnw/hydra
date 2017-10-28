package alarm

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) registryCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response =context.GetStandardResponse()
	title := ctx.Input.GetArgsValue("title", "注册中心服务")
	msg := ctx.Input.GetArgsValue("msg", "注册中心服务:@url当前数量:@current")
	path, err := ctx.Input.GetArgsByName("path")
	if err != nil {
		return
	}
	minValue, err := ctx.Input.GetArgsIntValue("min")
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
	tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "registry", tf, value)
	response.SetError(st, err)
	return
}
