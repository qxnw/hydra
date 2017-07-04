package rpc_proxy

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/types"
)

type rpcProxy struct {
	domain     string
	serverName string
	serverType string
	services   []string
	invoker    *rpc.RPCInvoker
}

func newRPCProxy() *rpcProxy {
	return &rpcProxy{
		services: make([]string, 0, 1),
	}
}

func (s *rpcProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	s.invoker = ctx.Invoker
	return s.services, nil

}
func (s *rpcProxy) Close() error {
	s.invoker.Close()
	return nil
}
func (s *rpcProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {

	input := make(map[string]string)
	ctx.GetInput().Each(func(k string, v string) {
		input[k] = v
	})
	input["__body"] = ctx.GetBody()
	input["hydra_sid"] = ctx.GetExt()["hydra_sid"].(string)
	status, result, params, err := s.invoker.Request(service, input, true)
	if err != nil {
		err = fmt.Errorf("engine:rpc_proxy.%v,status：%v,%v", err, status, result)
	}
	return &context.Response{Status: status, Content: result, Params: types.GetIMap(params)}, err
}
func (s *rpcProxy) Has(shortName, fullName string) (err error) {
	_, err = s.invoker.GetClientFromPool(fullName)
	return err
}

type rpcProxyResolver struct {
}

func (s *rpcProxyResolver) Resolve() engine.IWorker {
	return newRPCProxy()
}

func init() {
	engine.Register("rpc", &rpcProxyResolver{})
}
