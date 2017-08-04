package collect

import (
	"net"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) tcpCollect(ctx *context.Context) (r string, st int, err error) {
	title := ctx.GetArgValue("title", "TCP服务器")
	msg := ctx.GetArgValue("msg", "TCP服务器地址:@url")
	host, err := ctx.GetArgByName("host")
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
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err = s.checkAndSave(ctx, "tcp", tf, result)
	return
}