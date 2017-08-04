package collect

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) cpuCollect(ctx *context.Context) (r string, st int, err error) {
	title := ctx.GetArgValue("title", "服务器CPU负载")
	msg := ctx.GetArgValue("msg", "@host服务器CPU负载:@current")
	maxValue, err := ctx.GetArgFloat64Value("max")
	if err != nil {
		return
	}
	cpuInfo := cpu.GetInfo()
	value := 1
	if cpuInfo.UsedPercent < maxValue {
		value = 0
	}
	tf := transform.New()

	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", fmt.Sprintf("%.2f", cpuInfo.UsedPercent))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err = s.checkAndSave(ctx, "cpu", tf, value)
	return
}
