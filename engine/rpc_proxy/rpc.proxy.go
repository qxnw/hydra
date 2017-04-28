package rpc_proxy

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/transform"
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
		services: make([]string, 0, 16),
	}
}

func (s *rpcProxy) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	s.invoker = invoker
	return []string{"*"}, nil

}
func (s *rpcProxy) Close() error {
	s.invoker.Close()
	return nil
}
func (s *rpcProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {

	input := map[string]string{}
	if ctx.Input.Input == nil {
		return &context.Response{Status: 500}, fmt.Errorf("输入参数为空:%s", service)
	}
	if d, ok := ctx.Input.Input.(transform.ITransformGetter); ok {
		d.Each(func(k string, v string) {
			input[k] = v
		})
	}
	input["hydra_sid"] = ctx.Ext["hydra_sid"].(string)
	status, result, err := s.invoker.Request(service, input, true)
	return &context.Response{Status: status, Content: result}, err
}
func (s *rpcProxy) Has(service string) (err error) {
	_, err = s.invoker.Get(service)
	return err
	//return nil
}

type rpcProxyResolver struct {
}

func (s *rpcProxyResolver) Resolve() engine.IWorker {
	return newRPCProxy()
}

func init() {
	engine.Register("rpc", &rpcProxyResolver{})
}
