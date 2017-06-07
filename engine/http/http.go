package http

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/types"
)

type httpProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	encrypts        []string
	serviceHandlers map[string]func(*context.Context) (string, int, error)
}

func newHTTPProxy() *httpProxy {
	r := &httpProxy{
		services: make([]string, 0, 1),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, int, error))
	r.serviceHandlers["/http/handle"] = r.httpHandle
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	r.encrypts = []string{"md5", "base64", "rsa/sha1", "rsa/md5", "aes", "des"}
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
	if err = s.Has(service, service); err != nil {
		return
	}
	content, t, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:http.%v", err)
		return &context.Response{Status: types.DecodeInt(t, 0, 500)}, err
	}
	return &context.Response{Status: types.DecodeInt(t, 0, 200), Content: content}, nil
}

func (s *httpProxy) Has(shortName, fullName string) (err error) {
	for _, v := range s.services {
		if v == shortName {
			return nil
		}
	}
	return fmt.Errorf("engine:http:不存在服务:%s", shortName)
}

type httpProxyResolver struct {
}

func (s *httpProxyResolver) Resolve() engine.IWorker {
	return newHTTPProxy()
}

func init() {
	engine.Register("http", &httpProxyResolver{})
}
