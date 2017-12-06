package alarm

import (
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) netConnectCountCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	title := ctx.Input.GetArgsValue("title", "网络连接数")
	platform := ctx.Input.GetArgsValue("platform", "----")
	msg := ctx.Input.GetArgsValue("msg", "@host服务器网络连接数:@current")
	maxValue, err := ctx.Input.GetArgsIntValue("max")
	if err != nil {
		return
	}
	ncc, err := s.getNetConnectCount()
	if err != nil {
		return
	}
	value := 1
	if ncc < maxValue {
		value = 0
	}
	tf := transform.New()
	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", strconv.Itoa(ncc))
	tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	tf.Set("platform", platform)
	st, err := s.checkAndSave(ctx, "ncc", tf, value)
	response.SetError(st, err)
	return
}

func (s *collectProxy) getNetConnectCount() (v int, err error) {
	count, err := pipes.BashRun(`netstat -an|grep tcp|wc -l`)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
func (s *collectProxy) getMaxOpenFiles() (v int, err error) {
	count, err := pipes.BashRun("ulimit -n")
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return
}
