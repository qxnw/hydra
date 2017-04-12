package script

import (
	"fmt"
	"io/ioutil"

	"strings"

	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/file"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lua4go"
	"github.com/qxnw/lua4go/bind"
)

var (
	METHOD_NAME = []string{"request", "query", "delete", "update", "insert"}
)

type scriptPlugin struct {
	domain     string
	serverName string
	serverType string
	scriptPath string
	vm         *lua4go.LuaVM
	services   map[string]string
}

func newScriptPlugin() *scriptPlugin {
	return &scriptPlugin{
		services: make(map[string]string),
		vm:       lua4go.NewLuaVM(bind.NewDefault(), 1, 10, time.Second),
	}
}

func (s *scriptPlugin) Start(domain string, serverName string, serverType string) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	services = make([]string, 0, 0)
	path := fmt.Sprintf("%s/%s/%s/script", s.domain, s.serverName, s.serverType)
	p, err := file.GetAbs(path)
	if err != nil {
		return nil, err
	}
	svsDirNames, err := ioutil.ReadDir(p) //获取服务目录
	if err != nil {
		return
	}
	for _, v := range svsDirNames {
		if !v.IsDir() {
			continue
		}
		svsDirName := v.Name()
		svsPath := fmt.Sprintf("%s/%s", path, svsDirName)
		svsNames, err := ioutil.ReadDir(svsPath) //获取服务目录
		if err != nil {
			return nil, err
		}
		for _, sv := range svsNames {
			svName := sv.Name()
			for _, method := range METHOD_NAME {
				if strings.HasPrefix(svName, method+".") {
					filePath := fmt.Sprintf("%s/%s", svsPath, svName)
					if err = s.vm.PreLoad(filePath); err != nil {
						return nil, err
					}
					name := fmt.Sprintf("%s.%s", svsDirName, method)
					s.services[name] = filePath
					services = append(services, name)
				}
			}
		}

	}
	return
}

func (s *scriptPlugin) Close() error {
	s.vm.Close()
	return nil
}
func (s *scriptPlugin) Handle(name string, method string, service string, ctx *context.Context) (r *context.Response, err error) {
	svName := fmt.Sprintf("%s.%s", service, method)
	f, ok := s.services[svName]
	if !ok {
		return nil, fmt.Errorf("script plugin 未找到服务：%s", svName)
	}
	log := logger.GetSession(svName, ctx.Ext["hydra_sid"].(string))
	defer log.Close()
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
func init() {
	engine.Register("script", newScriptPlugin())
}
