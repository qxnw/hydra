package script

import (
	"fmt"
	"io/ioutil"
	"os"

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
		vm:          lua4go.NewLuaVM(bind.NewDefault(), 1, 100, time.Second*300), //引擎池5分钟不用则自动回收
	}
}

func (s *scriptWorker) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	s.invoker = invoker
	path := fmt.Sprintf("%s/servers/%s/%s/script", s.domain, s.serverName, s.serverType)
	p, err := file.GetAbs(path)
	if err != nil {
		return
	}
	s.services, err = s.findService(p, "")
	if err != nil {
		return
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
	fname = strings.ToUpper(s.getServiceName(name, parent))
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
	for _, method := range engine.METHOD_NAME {
		if strings.HasSuffix(svName, method+".lua") || strings.HasSuffix(svName, "."+method+".lua") {
			i := strings.LastIndex(svName, ".")
			return parent + svName[0:i]
		}
	}
	return ""
}
func (s *scriptWorker) Close() error {
	s.vm.Close()
	return nil
}
func (s *scriptWorker) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	fmt.Println("Handle.script:", service, s.domain, s.serverName, s.serverType)

	f, ok := s.srvsPathMap[service]
	if !ok {
		return &context.Response{Status: 404}, fmt.Errorf("script plugin 未找到服务：%s", service)
	}
	log := logger.GetSession(service, ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	ctx.Ext["__func_rpc_invoker_"] = s.invoker
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	result, m, err := s.vm.Call(f, input)
	if err != nil {
		return
	}
	data := make(map[string]interface{})
	for k, v := range m {
		data[k] = v
	}
	r = &context.Response{Status: 200, Params: data, Content: result[0]}
	return
}
func (s *scriptWorker) Has(service string) (err error) {
	for _, v := range s.services {
		if v == service {
			return nil
		}
	}
	return fmt.Errorf("不存在服务:%s", service)
}

type scriptResolver struct {
}

func (s *scriptResolver) Resolve() engine.IWorker {
	return newScriptWorker()
}

func init() {
	engine.Register("script", &scriptResolver{})
}
