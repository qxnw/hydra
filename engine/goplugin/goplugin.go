package goplugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"plugin"

	"strings"

	"sync"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra_plugin/plugins"
	"github.com/qxnw/lib4go/file"
)

var plugines map[string]plugins.PluginWorker
var mu sync.Mutex

//go build -buildmode=plugin
type goPluginWorker struct {
	domain     string
	serverName string
	serverType string
	scriptPath string
	srvPlugins map[string]plugins.PluginWorker
	services   []string
	invoker    *rpc.RPCInvoker
}

func newGoPluginWorker() *goPluginWorker {
	return &goPluginWorker{
		services:   make([]string, 0, 16),
		srvPlugins: make(map[string]plugins.PluginWorker),
	}
}

func (s *goPluginWorker) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	s.invoker = invoker
	path := fmt.Sprintf("%s/servers/%s/%s/go", s.domain, s.serverName, s.serverType)
	p, err := file.GetAbs(path)
	if err != nil {
		return
	}
	serviceNames, err := ioutil.ReadDir(path) //获取服务根目录
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	s.services = make([]string, 0, 16)
	pname := strings.Replace(serverName, ".", "_", -1)
	for _, v := range serviceNames {

		if v.IsDir() || !strings.HasPrefix(v.Name(), pname) || !strings.HasSuffix(v.Name(), ".so") {
			continue
		}
		pp := fmt.Sprintf("%s/%s", p, v.Name())
		rwrk, err := s.loadPlugin(v.Name(), pp)
		if err != nil {
			return nil, err
		}
		srvs := rwrk.GetServices()
		for _, v := range srvs {
			s.srvPlugins[v] = rwrk
		}
		for _, v := range srvs {
			s.services = append(s.services, strings.ToUpper(v))
		}
	}
	return s.services, nil
}
func (s *goPluginWorker) loadPlugin(name string, path string) (r plugins.PluginWorker, err error) {
	mu.Lock()
	defer mu.Unlock()
	if p, ok := plugines[name]; ok {
		return p, nil
	}
	pg, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("go pulgin加载失败:%s,err:%v", path, err)
	}
	work, err := pg.Lookup("GetWorker")
	if err != nil {
		return nil, fmt.Errorf("go pulgin未找到函数GetWorker:%s,err:%v", path, err)
	}
	wkr, ok := work.(func() plugins.PluginWorker)
	if !ok {
		return nil, fmt.Errorf("go pulgin的GetWorker函数必须为 func() PluginWorker 类型:%s", path)
	}
	rwrk := wkr()
	plugines[name] = rwrk
	return rwrk, nil
}
func (s *goPluginWorker) Close() error {
	return nil
}
func (s *goPluginWorker) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	f, ok := s.srvPlugins[service]
	if !ok {
		return &context.Response{Status: 404}, fmt.Errorf("go plugin 未找到服务：%s", service)
	}
	st, rs, err := f.Handle(svName, mode, service, ctx, s.invoker)
	return &context.Response{Status: st, Content: rs}, err
}
func (s *goPluginWorker) Has(service string) error {
	for _, v := range s.services {
		if v == service {
			return nil
		}
	}
	return fmt.Errorf("不存在服务:%s", service)
}

type goResolver struct {
}

func (s *goResolver) Resolve() engine.IWorker {
	return newGoPluginWorker()
}

func init() {
	plugines = make(map[string]plugins.PluginWorker)
	engine.Register("go", &goResolver{})
}
