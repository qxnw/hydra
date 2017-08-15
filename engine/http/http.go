package http

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type httpProxy struct {
	ctx             *engine.EngineContext
	services        []string
	encrypts        []string
	serviceHandlers map[string]context.WHandlerFunc
}

func newHTTPProxy() *httpProxy {
	r := &httpProxy{
		services: make([]string, 0, 1),
	}
	r.serviceHandlers = make(map[string]context.WHandlerFunc)
	r.serviceHandlers["/http/handle"] = r.httpHandle
	r.serviceHandlers["/http/redirect"] = r.httpRedirectHandle
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	r.encrypts = []string{"md5", "base64", "rsa/sha1", "rsa/md5", "aes", "des"}
	return r
}

func (s *httpProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	return s.services, nil
}
func (s *httpProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取 receiver,subject,content
//从args参数中获取 mail
//配置文件格式:{"smtp":"smtp.exmail.qq.com:25", "sender":"yanglei@100bm.cn","password":"12333"}

func (s *httpProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:http.%v", err)
	}
	return
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
