package ssm

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/types"
)

type smsProxy struct {
	ctx             *engine.EngineContext
	services        []string
	serviceHandlers map[string]func(*context.Context) (string, int, error)
}

func newSmsProxy() *smsProxy {
	p := &smsProxy{
		services:        make([]string, 0, 4),
		serviceHandlers: make(map[string]func(*context.Context) (string, int, error)),
	}
	p.serviceHandlers["/ssm/ytx/send"] = p.ytxSend
	p.serviceHandlers["/ssm/wx/send"] = p.wxSend
	p.serviceHandlers["/ssm/wx0/send"] = p.wxSend0
	p.serviceHandlers["/ssm/wx1/send"] = p.wxSend1
	p.serviceHandlers["/ssm/email/send"] = p.sendMail
	for k := range p.serviceHandlers {
		p.services = append(p.services, k)
	}
	return p
}

func (s *smsProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
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
		return &context.Response{Status: types.DecodeInt(st, 0, 500)}, fmt.Errorf("engine:ssm.%v", err)
	}
	return &context.Response{Status: types.DecodeInt(st, 0, 200), Content: content}, nil
}
func (s *smsProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:ssm.不存在服务:%s", shortName)
}

type ytxProxyResolver struct {
}

func (s *ytxProxyResolver) Resolve() engine.IWorker {
	return newSmsProxy()
}

func init() {
	engine.Register("ssm", &ytxProxyResolver{})
}
