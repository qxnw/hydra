package log

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type logProxy struct {
	domain          string
	serverName      string
	serverType      string
	services        []string
	serviceHandlers map[string]context.SHandlerFunc
}

func newLogProxy() *logProxy {
	r := &logProxy{
		services: make([]string, 0, 1),
	}
	r.serviceHandlers = make(map[string]context.SHandlerFunc)
	r.serviceHandlers["/log/error"] = r.logFileErrorHandle
	r.serviceHandlers["/log/info"] = r.logFileInfoHandle
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *logProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *logProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取 receiver,subject,content
//从args参数中获取 mail
//配置文件格式:{"smtp":"smtp.exmail.qq.com:25", "sender":"yanglei@100bm.cn","password":"12333"}

func (s *logProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	r, err = s.serviceHandlers[service](svName, mode, service, ctx)
	if err != nil {
		err = fmt.Errorf("engine:log.%v", err)
	}
	return
}

func (s *logProxy) Has(shortName, fullName string) (err error) {
	for _, v := range s.services {
		if v == shortName {
			return nil
		}
	}
	return fmt.Errorf("engine:log:不存在服务:%s", shortName)
}

type logProxyResolver struct {
}

func (s *logProxyResolver) Resolve() engine.IWorker {
	return newLogProxy()
}

func init() {
	engine.Register("log", &logProxyResolver{})
}
