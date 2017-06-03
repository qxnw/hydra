package script

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"strings"

	"time"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/file"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lua4go"
	"github.com/qxnw/lua4go/bind"
)

type scriptWorker struct {
	domain      string
	serverName  string
	serverType  string
	scriptPath  string
	vm          *lua4go.LuaVM
	srvsPathMap map[string]string
	services    []string
	invoker     *rpc.RPCInvoker
}

func newScriptWorker() *scriptWorker {
	return &scriptWorker{
		srvsPathMap: make(map[string]string),
		services:    make([]string, 0, 16),
	}
}
func (w *scriptWorker) getPathPrefix() string {
	prefix := ""
	if runtime.GOOS == "windows" {
		p := os.Args[0]
		prefix = string(p[0]) + ":"
	}
	return prefix
}
func (s *scriptWorker) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	s.invoker = ctx.Invoker

	path := fmt.Sprintf("%s%s/servers/%s/%s/script", s.getPathPrefix(), s.domain, s.serverName, s.serverType)
	p, err := file.GetAbs(path)
	if err != nil {
		return
	}
	opts := make([]lua4go.Option, 0, 4)
	opts = append(opts, lua4go.WithMinSize(1))
	opts = append(opts, lua4go.WithMaxSize(99))
	opts = append(opts, lua4go.WithTimeout(time.Second*300)) //引擎池5分钟不用则自动回收
	if engine.IsDebug {
		opts = append(opts, lua4go.WithWatchScript(time.Second))
	}
	s.vm = lua4go.NewLuaVM(bind.NewDefault([]string{p, p + "/xlib", p + "/lib"}...), opts...)
	s.services, err = s.findService(p, "")
	if err != nil {
		s.vm.Close()
		return
	}
	if len(s.services) == 0 {
		s.vm.Close()
	}
	return s.services, nil
}
func (s *scriptWorker) findService(path string, parent string) (services []string, err error) {
	services = make([]string, 0, 0)
	serviceNames, err := ioutil.ReadDir(path) //获取服务根目录
	if err != nil && !os.IsNotExist(err) {
		return nil, nil
	}
	for _, v := range serviceNames {
		//当前目录中搜索服务
		rootServiceName := v.Name()
		rootServicePath := fmt.Sprintf("%s/%s", path, rootServiceName)
		if !v.IsDir() {
			svName, err := s.loadService(rootServiceName, parent+"/", path)
			if err != nil {
				return nil, err
			}
			if svName != "" {
				services = append(services, svName)
			}
			continue
		}
		srvs, err := s.findService(rootServicePath, parent+"/"+rootServiceName)
		if err != nil {
			return nil, err
		}
		services = append(services, srvs...)

	}
	return services, nil
}
func (s *scriptWorker) loadService(name string, parent string, root string) (fname string, err error) {
	fname = strings.ToLower(s.getServiceName(name, parent))
	if fname == "" {
		return "", nil
	}
	filePath := fmt.Sprintf("%s/%s", root, name)
	if err = s.vm.PreLoad(filePath); err != nil {
		return
	}
	s.srvsPathMap[fname] = filePath
	return
}
func (s *scriptWorker) getServiceName(svName string, parent string) string {
	for _, method := range engine.EXCLUDE {
		if (!strings.HasSuffix(svName, ".lua") && !strings.HasSuffix(svName, ".luac")) || strings.Contains(svName, method) || strings.Contains(parent, method) {
			return ""
		}
	}
	i := strings.LastIndex(svName, ".")
	return parent + svName[0:i]

}
func (s *scriptWorker) Close() error {
	s.vm.Close()
	return nil
}
func (s *scriptWorker) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {

	f, ok := s.srvsPathMap[svName]
	if !ok {
		return &context.Response{Status: 404}, fmt.Errorf("script plugin 未找到服务：%s", svName)
	}
	log := logger.GetSession(svName, ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	ctx.Ext["__func_rpc_invoker_"] = s.invoker
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	result, m, err := s.vm.Call(f, input)
	if err != nil {
		err = fmt.Errorf("engine:script:%v", err)
		return
	}
	data := make(map[string]interface{})
	for k, v := range m {
		data[k] = v
	}
	r = &context.Response{Status: 200, Params: data, Content: result[0]}
	return
}
func (s *scriptWorker) Has(shortName, fullName string) (err error) {
	for _, v := range s.services {
		if v == shortName {
			return nil
		}
	}
	return fmt.Errorf("engine:script不存在服务:%s", shortName)
}

type scriptResolver struct {
}

func (s *scriptResolver) Resolve() engine.IWorker {
	return newScriptWorker()
}

func init() {
	engine.Register("script", &scriptResolver{})
}
