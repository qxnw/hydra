package ssm

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type smsProxy struct {
	ctx             *engine.EngineContext
	services        []string
	serviceHandlers map[string]context.SHandlerFunc
}

func newSmsProxy() *smsProxy {
	p := &smsProxy{
		services:        make([]string, 0, 4),
		serviceHandlers: make(map[string]context.SHandlerFunc),
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
func (s *smsProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:ssm.%v", err)
	}
	return
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
