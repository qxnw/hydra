package collect

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (s *collectProxy) httpCollect(ctx *context.Context) (r string, st int, err error) {
	title := ctx.GetArgValue("title", "HTTP服务器")
	msg := ctx.GetArgValue("msg", "HTTP服务器地址:@url,请求响应码@current")
	uri, err := ctx.GetArgByName("url")
	if err != nil {
		return
	}
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("http请求参数url配置有误:%v", uri)
		return
	}
	client := http.NewHTTPClient()
	_, t, err := client.Get(uri)
	value := types.DecodeInt(t, 200, 0, 1)
	tf := transform.New()
	tf.Set("host", u.Host)
	tf.Set("url", uri)
	tf.Set("value", strconv.Itoa(value))
	tf.Set("current", strconv.Itoa(t))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("unq", tf.Translate("@url"))
	tf.Set("title", tf.Translate(title))
	tf.Set("msg", tf.Translate(msg))
	st, err = s.checkAndSave(ctx, "http", tf, value)
	return
}