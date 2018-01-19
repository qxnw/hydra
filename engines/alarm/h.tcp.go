package alarm

import (
	"net"
	"strconv"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

//TCPStatusCollect tcp状态收集
func TCPStatusCollect(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		title := ctx.Input.GetArgsValue("title", "TCP服务器")
		msg := ctx.Input.GetArgsValue("msg", "TCP服务器地址:@url")
		platform := ctx.Input.GetArgsValue("platform", "----")
		host, err := ctx.Input.GetArgsByName("host")
		if err != nil {
			return
		}
		conn, err := net.DialTimeout("tcp", host, time.Second)
		if err == nil {
			conn.Close()
		}
		result := types.DecodeInt(err, nil, 0, 1)
		tf := transform.NewMap(map[string]string{})
		tf.Set("host", host)
		tf.Set("url", host)
		tf.Set("value", strconv.Itoa(result))
		tf.Set("level", ctx.Input.GetArgsValue("level", "1"))
		tf.Set("group", ctx.Input.GetArgsValue("group", "D"))
		tf.Set("time", time.Now().Format("20060102150405"))
		tf.Set("unq", tf.Translate("@host"))
		tf.Set("title", tf.Translate(title))
		tf.Set("msg", tf.Translate(msg))
		tf.Set("platform", platform)
		st, err := checkAndSave(c, tf, result, "tcp")
		response.SetError(st, err)
		return
	}
}
