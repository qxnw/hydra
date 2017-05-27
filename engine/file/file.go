package file

import (
	"fmt"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type fileProxy struct {
	domain     string
	serverName string
	serverType string
	services   []string

	serviceHandlers map[string]func(*context.Context) (string, error)
}

func newFileProxy() *fileProxy {
	r := &fileProxy{
		services: make([]string, 0, 8),
	}
	r.serviceHandlers = make(map[string]func(*context.Context) (string, error), 8)
	r.serviceHandlers["/file/upload"] = r.saveFileFromHTTPRequest
	for k := range r.serviceHandlers {
		r.services = append(r.services, k)
	}
	return r
}

func (s *fileProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	services = s.services
	return

}
func (s *fileProxy) Close() error {
	return nil
}
func (s *fileProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if err = s.Has(service, service); err != nil {
		return
	}
	content, err := s.serviceHandlers[service](ctx)
	if err != nil {
		err = fmt.Errorf("engine:file %v", err)
		return &context.Response{Status: 500}, err
	}
	return &context.Response{Status: 200, Content: content}, nil
}
func (s *fileProxy) Has(shortName, fullName string) (err error) {
	if _, ok := s.serviceHandlers[shortName]; ok {
		return nil
	}
	return fmt.Errorf("engine:file不存在服务:%s", shortName)
}

type fileProxyResolver struct {
}

func (s *fileProxyResolver) Resolve() engine.IWorker {
	return newFileProxy()
}

func init() {
	engine.Register("file", &fileProxyResolver{})
}
