package monitor

import (
	"github.com/qxnw/hydra/context"
	xnet "github.com/qxnw/lib4go/net"
)

func (s *monitorProxy) registryCollect(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	path, err := ctx.Input.GetArgsByName("path")
	if err != nil {
		return
	}
	data, _, err := s.registry.GetChildren(path)
	if err != nil {
		return
	}
	ip := xnet.GetLocalIPAddress(ctx.Input.GetArgsValue("mask", ""))
	err = updateRegistryStatus(ctx, int64(len(data)), "server", ip, "path", path)
	response.SetError(0, err)
	return
}
