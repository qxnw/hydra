package monitor

import (
	"fmt"
	"net/url"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/types"
)

func (s *monitorProxy) httpCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	uri, err := ctx.Input.GetArgsByName("url")
	if err != nil {
		return
	}
	_, err = url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("http请求参数url配置有误:%v", uri)
		return
	}
	client := http.NewHTTPClient()
	_, t, err := client.Get(uri)
	value := types.DecodeInt(t, 200, 0, 1)
	ip := net.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateHTTPStatus(ctx, int64(value), "server", ip, "url", uri)
	response.SetError(0, err)
	return
}
