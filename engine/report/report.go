package report

import (
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/types"
)

type reportProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	invoker         *rpc.Invoker
	ctx             *engine.EngineContext
	serviceHandlers map[string]func(*context.Context) (string, int, error)
}

func newReportProxy() *reportProxy {
	p := &reportProxy{
		services:        make([]string, 0, 4),
		serviceHandlers: make(map[string]func(*context.Context) (string, int, error)),
	}
	p.serviceHandlers["/report/sql/query"] = p.sqlQueryHandle
	for k := range p.serviceHandlers {
		p.services = append(p.services, k)
	}
	return p
}

func (s *reportProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	return s.services, nil

}
func (s *reportProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取: mobile,data
//从args参数中获取:mail
//ytx配置文件内容：见ytx.go
func (s *reportProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	content, st, err := s.serviceHandlers[service](ctx)
	if err != nil {
		return &context.Response{Status: types.DecodeInt(st, 0, 500)}, fmt.Errorf("engine:report.%v", err)
	}
	return &context.Response{Status: types.DecodeInt(st, 0, 200), Content: content}, nil
}
func (s *reportProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:report.不存在服务:%s", shortName)
}

type reportProxyResolver struct {
}

func (s *reportProxyResolver) Resolve() engine.IWorker {
	return newReportProxy()
}

func init() {
	engine.Register("report", &reportProxyResolver{})
}
