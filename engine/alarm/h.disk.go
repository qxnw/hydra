package alarm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) diskCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	title := ctx.Input.GetArgsValue("title", "服务器disk使用率")
	msg := ctx.Input.GetArgsValue("msg", "@host服务器disk使用率:@current")
	maxValue, err := ctx.Input.GetArgsFloat64Value("max")
	if err != nil {
		return
	}
	diskInfo := disk.GetInfo()
	result := 1
	if diskInfo.UsedPercent < maxValue {
		result = 0
	}
	tf := transform.New()
	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("current", fmt.Sprintf("%.2f", diskInfo.UsedPercent))
	tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "disk", tf, result)
	response.SetError(st, err)
	return
}