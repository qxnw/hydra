package monitor

import (
	"net"
	"time"

	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/types"
)

func (s *monitorProxy) tcpCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	host, err := ctx.Input.GetArgsByName("host")
	if err != nil {
		return
	}
	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err == nil {
		conn.Close()
	}
	value := types.DecodeInt(err, nil, 0, 1)
	ip := xnet.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateTCPStatus(ctx, ctx.Input.GetArgsValue("influxdb", "alarm"), int64(value), "server", ip, "host", host)
	response.SetError(0, err)
	return
}
