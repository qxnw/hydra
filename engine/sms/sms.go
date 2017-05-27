package sms

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type smsProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	invoker         *rpc.RPCInvoker
	serviceHandlers map[string]func(*context.Context) (string, int, error)
}

func newSmsProxy() *smsProxy {
	p := &smsProxy{
		services:        make([]string, 0, 1),
		serviceHandlers: make(map[string]func(*context.Context) (string, int, error)),
	}
	p.serviceHandlers["/sms/ytx/send"] = p.ytxSend
	for k := range p.serviceHandlers {
		p.services = append(p.services, k)
	}
	return p
}

func (s *smsProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	s.invoker = ctx.Invoker
	return s.services, nil

}
func (s *smsProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取: mobile,data
//从args参数中获取:mail
//ytx配置文件内容：见ytx.go
func (s *smsProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	content, st, err := s.serviceHandlers[service](ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: st, Content: content}, nil
}
func (s *smsProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("不存在服务:%s", shortName)
}

type ytxProxyResolver struct {
}

func (s *ytxProxyResolver) Resolve() engine.IWorker {
	return newSmsProxy()
}

func init() {
	engine.Register("sms", &ytxProxyResolver{})
}
