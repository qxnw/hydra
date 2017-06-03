package http

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/utility"
)

type httpProxy struct {
	domain     string
	serverName string
	serverType string
	services   []string
}

func newHTTPProxy() *httpProxy {
	r := &httpProxy{
		services: make([]string, 0, 1),
	}
	return r
}

func (s *httpProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *httpProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取 receiver,subject,content
//从args参数中获取 mail
//配置文件格式:{"smtp":"smtp.exmail.qq.com:25", "sender":"yanglei@100bm.cn","password":"12333"}

func (s *httpProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {

	content, t, err := s.httpHandle(service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:http.%v", err)
		return &context.Response{Status: utility.EqualAndSet(t, 0, 500)}, err
	}
	return &context.Response{Status: utility.EqualAndSet(t, 0, 200), Content: content}, nil
}

func (s *httpProxy) Has(shortName, fullName string) (err error) {
	return nil
}

type httpProxyResolver struct {
}

func (s *httpProxyResolver) Resolve() engine.IWorker {
	return newHTTPProxy()
}

func init() {
	engine.Register("http", &httpProxyResolver{})
}
