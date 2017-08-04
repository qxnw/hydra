package collect

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) registryCollect(ctx *context.Context) (r string, st int, err error) {
	title := ctx.GetArgValue("title", "注册中心服务")
	msg := ctx.GetArgValue("msg", "注册中心服务:@url,当前数量:@current")
	path, err := ctx.GetArgByName("path")
	if err != nil {
		return
	}
	minValue, err := ctx.GetArgIntValue("min")
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
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err = s.checkAndSave(ctx, "registry", tf, value)
	return
}
