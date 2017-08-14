package goplugin

import (
	"fmt"

	"strings"

	"sync"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
)

var plugines map[string]context.Worker
var mu sync.Mutex

//go build -buildmode=plugin
type goPluginWorker struct {
	ctx        *engine.EngineContext
	scriptPath string
	srvPlugins map[string]context.Worker
	services   cmap.ConcurrentMap
	path       []string
	isClose    bool
}

func newGoPluginWorker() *goPluginWorker {
	return &goPluginWorker{
		services:   cmap.New(2),
		srvPlugins: make(map[string]context.Worker),
		path:       make([]string, 0, 2),
	}
}

func (s *goPluginWorker) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.ctx = ctx
	s.path = append(s.path, fmt.Sprintf("./%s_%s.so", ctx.ServerName, ctx.ServerType))
	s.path = append(s.path, fmt.Sprintf("./%s.so", ctx.ServerName))
	for _, p := range s.path {
		rwrk, err := s.loadPlugin(p)
		if err != nil {
			return nil, err
		}
		if rwrk != nil {
			srvs := rwrk.GetServices()
			for _, v := range srvs {
				svName := strings.ToLower(v)
				s.srvPlugins[svName] = rwrk
				s.services.SetIfAbsent(svName, svName)
				services = append(services, svName)
			}
		}

	}

	return services, nil

}
func (s *goPluginWorker) Close() error {
	s.isClose = true
	for _, v := range s.srvPlugins {
		v.Close()
	}
	return nil
}

//Handle 从bin目录下获取当前应用匹配的动态库，并返射加载服务
func (s *goPluginWorker) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	if s.isClose {
		return &context.Response{Status: 520}, fmt.Errorf("engine:goplugin.服务已关闭：%s", svName)
	}
	f, ok := s.srvPlugins[svName]
	if !ok {
		return &context.Response{Status: 404}, fmt.Errorf("engine:goplugin.未找到服务：%s", svName)
	}
	return f.Handle(svName, mode, service, ctx)
}
func (s *goPluginWorker) Has(shortName, fullName string) error {
	if s.services.Count() == 0 {
		return fmt.Errorf("engine:goplugin.在目录:%v中未找到go插件", s.path)
	}
	if s.services.Has(shortName) {
		return nil
	}
	return fmt.Errorf("engine:goplugin.不存在服务:%s:%d", shortName, s.services.Count())

}

type goResolver struct {
}

func (s *goResolver) Resolve() engine.IWorker {
	return newGoPluginWorker()
}

func init() {
	plugines = make(map[string]context.Worker)
	engine.Register("go", &goResolver{})
}
