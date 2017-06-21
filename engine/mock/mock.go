package mock

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/types"
)

type mockProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	serviceHandlers map[string]func(*context.Context) (string, int, map[string]interface{}, error)
}

func newMockProxy() *mockProxy {
	r := &mockProxy{
		services: make([]string, 0, 1),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, int, map[string]interface{}, error))
	r.serviceHandlers["/mock/raw/request"] = r.rawMockHandle
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *mockProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *mockProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取 receiver,subject,content
//从args参数中获取 mail
//配置文件格式:{"smtp":"smtp.exmail.qq.com:25", "sender":"yanglei@100bm.cn","password":"12333"}

func (s *mockProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	content, t, header, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:http.%v", err)
		return &context.Response{Status: types.DecodeInt(t, 0, 500)}, err
	}
	return &context.Response{Status: types.DecodeInt(t, 0, 200), Content: content, Params: header}, nil

}

func (s *mockProxy) Has(shortName, fullName string) (err error) {
	for _, v := range s.services {
		if v == shortName {
			return nil
		}
	}
	return fmt.Errorf("engine:http:不存在服务:%s", shortName)
}

type mockProxyResolver struct {
}

func (s *mockProxyResolver) Resolve() engine.IWorker {
	return newMockProxy()
}

func init() {
	engine.Register("mock", &mockProxyResolver{})
}