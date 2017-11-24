package alarm

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/pipes"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) nginxErrorCountCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	title := ctx.Input.GetArgsValue("title", "nginx错误数")
	msg := ctx.Input.GetArgsValue("msg", "@host服务器nginx错误数:@current(@ct)")
	maxValue, err := ctx.Input.GetArgsIntValue("max")
	if err != nil {
		return
	}
	current, ct, err := s.getNginxErrorCount()
	if err != nil {
		return
	}
	value := 1
	if current < maxValue {
		value = 0
	}
	tf := transform.New()
	tf.Set("host", net.LocalIP)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", strconv.Itoa(current))
	tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
	tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
	tf.Set("ct", ct)
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@host"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err := s.checkAndSave(ctx, "nginx-error", tf, value)
	response.SetError(st, err)
	return
}
func (s *collectProxy) getNginxErrorCount() (m int, tm string, err error) {
	tm = time.Now().Add(-1 * time.Minute).Format("15:04")
	cmd1 := exec.Command("cat", "/usr/local/nginx/logs/error.log")
	cmd2 := exec.Command("grep", fmt.Sprintf(`"%s:"`, tm))
	cmd3 := exec.Command("wc", "-l")
	cmds := []*exec.Cmd{cmd1, cmd2, cmd3}
	count, err := pipes.Run(cmds)
	if err != nil {
		err = fmt.Errorf("命令执行失败:%v-%s", err, tm)
		return
	}

	v, err := strconv.Atoi(count)
	if err != nil {
		err = fmt.Errorf("%v:%s", err, count)
		return
	}
	m = v / 60
	return
}
