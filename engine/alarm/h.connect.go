package alarm

import (
	"os/exec"
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
	st, err := s.checkAndSave(ctx, "ncc", tf, value)
	response.SetError(st, err)
	return
}

func (s *collectProxy) getNetConnectCount() (v int, err error) {
	cmd1 := exec.Command("netstat", "-an")
	cmd2 := exec.Command("grep", "tcp")
	cmd3 := exec.Command("wc", "-l")
	cmds := []*exec.Cmd{cmd1, cmd2, cmd3}
	count, err := pipes.Run(cmds)
	if err != nil {
		return
	}
	v, err = strconv.Atoi(count)
	return
}
func (s *collectProxy) getMaxOpenFiles() (v int, err error) {
	cmd1 := exec.Command("ulimit ", "-n")
	cmds := []*exec.Cmd{cmd1}
	count, err := pipes.Run(cmds)
	if err != nil {
		return
	}
	v, _ = strconv.Atoi(count)
	return v, nil
}
