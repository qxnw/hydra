package collect

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/memory"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) memCollect(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	title := ctx.Input.GetArgValue("title", "服务器memory使用率")
	msg := ctx.Input.GetArgValue("msg", "@host服务器memory使用率:@current")
	maxValue, err := ctx.Input.GetArgFloat64Value("max")
	if err != nil {
		return
	}
	memoryInfo := memory.GetInfo()
	value := 1
	if memoryInfo.UsedPercent < maxValue {
		value = 0
	}
	tf := transform.New()
	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", fmt.Sprintf("%.2f", memoryInfo.UsedPercent))
	tf.Set("level", ctx.Input.GetArgValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "mem", tf, value)
	response.Set(st, err)
	return
}
