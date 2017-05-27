package goplugin

import (
	"fmt"
	"os"
	"plugin"

	"strings"

	"sync"

	"github.com/qxnw/goplugin"
	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/file"
)

var plugines map[string]goplugin.PluginWorker
var mu sync.Mutex

//go build -buildmode=plugin
type goPluginWorker struct {
	domain     string
	serverName string
	serverType string
	scriptPath string
	srvPlugins map[string]goplugin.PluginWorker
	services   cmap.ConcurrentMap
	invoker    *rpc.RPCInvoker
	path       []string
}

func newGoPluginWorker() *goPluginWorker {
	return &goPluginWorker{
		services:   cmap.New(),
		srvPlugins: make(map[string]goplugin.PluginWorker),
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
	//fmt.Println("s.path:", s.path)
	for _, p := range s.path {
		rwrk, err := s.loadPlugin(p)
		if err != nil {
			return nil, err
		}
		if rwrk == nil {
			continue
		}
		srvs := rwrk.GetServices()
		for _, v := range srvs {
			svName := strings.ToLower(v)
			s.srvPlugins[svName] = rwrk
			s.services.SetIfAbsent(svName, svName)
			services = append(services, svName)
		}
	}

	return services, nil

}
func (s *goPluginWorker) loadPlugin(p string) (r goplugin.PluginWorker, err error) {
	mu.Lock()
	defer mu.Unlock()
	path, err := file.GetAbs(p)
	if err != nil {
		return
	}
	if p, ok := plugines[path]; ok {
		return p, nil
	}
	if _, err = os.Lstat(p); err != nil && os.IsNotExist(err) {
		return nil, nil
	}

	pg, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("go pulgin加载失败:%s,err:%v", path, err)
	}
	work, err := pg.Lookup("GetWorker")
	if err != nil {
		return nil, fmt.Errorf("go pulgin未找到函数GetWorker:%s,err:%v", path, err)
	}
	wkr, ok := work.(func() goplugin.PluginWorker)
	if !ok {
		return nil, fmt.Errorf("go pulgin的GetWorker函数必须为 func() PluginWorker 类型:%s", path)
	}
	rwrk := wkr()
	plugines[path] = rwrk
	return rwrk, nil
}
func (s *goPluginWorker) Close() error {
	return nil
}

//Handle 从bin目录下获取当前应用匹配的动态库，并返射加载服务
func (s *goPluginWorker) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	f, ok := s.srvPlugins[svName]
	if !ok {
		return &context.Response{Status: 404}, fmt.Errorf("go plugin 未找到服务：%s", svName)
	}
	st, rs, pa, err := f.Handle(svName, mode, service, ctx, s.invoker)
	return &context.Response{Status: st, Content: rs, Params: pa}, err
}
func (s *goPluginWorker) Has(shortName, fullName string) error {
	if s.services.Count() == 0 {
		return fmt.Errorf("在目录:%v中未找到go插件", s.path)
	}
	if s.services.Has(shortName) {
		return nil
	}
	return fmt.Errorf("不存在服务:%s:%d", shortName, s.services.Count())

}

type goResolver struct {
}

func (s *goResolver) Resolve() engine.IWorker {
	return newGoPluginWorker()
}

func init() {
	plugines = make(map[string]goplugin.PluginWorker)
	engine.Register("go", &goResolver{})
}
