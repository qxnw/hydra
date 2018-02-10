package monitor

import (
	"net"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/types"
)

//CollectTCPStatus 收集tcp状态
func CollectTCPStatus(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		host, err := ctx.Request.Setting.Get("host")
		if err != nil {
			return
		}
		conn, err := net.DialTimeout("tcp", host, time.Second)
		if err == nil {
			conn.Close()
		}
		value := types.DecodeInt(err, nil, 0, 1)
		ip := xnet.GetLocalIPAddress(ctx.Request.Setting.GetString("mask", ""))
		err = updateTCPStatus(c, ctx, int64(value), "server", ip, "host", host)
		response.SetContent(0, err)
		return
	}
}
