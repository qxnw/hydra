package goplugin

import (
	"fmt"

	"strings"

	"sync"

	"github.com/qxnw/goplugin"
	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/types"
)

var plugines map[string]goplugin.Worker
var mu sync.Mutex

//go build -buildmode=plugin
type goPluginWorker struct {
	domain     string
	serverName string
	serverType string
	scriptPath string
	srvPlugins map[string]goplugin.Worker
	services   cmap.ConcurrentMap
	invoker    *rpc.Invoker
	path       []string
	isClose    bool
}

func newGoPluginWorker() *goPluginWorker {
	return &goPluginWorker{
		services:   cmap.New(2),
		srvPlugins: make(map[string]goplugin.Worker),
		path:       make([]string, 0, 2),
	}
}

func (s *goPluginWorker) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	s.invoker = ctx.Invoker
	s.path = append(s.path, fmt.Sprintf("./%s_%s.so", s.serverName, s.serverType))
	s.path = append(s.path, fmt.Sprintf("./%s.so", s.serverName))
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
	st, rs, pa, err := f.Handle(svName, mode, service, ctx, s.invoker)
	st = types.DecodeInt(err != nil && (st == 0 || st == 200), true, 500, st)
	return &context.Response{Status: st, Content: rs, Params: pa}, err
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
	plugines = make(map[string]goplugin.Worker)
	engine.Register("go", &goResolver{})
}
