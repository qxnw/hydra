package script

import (
	"fmt"
	"io/ioutil"

	"strings"

	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/file"
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
	services map[string]string
}

func newScriptPlugin(domain string, serverName string, serverType string) *scriptPlugin {
	return &scriptPlugin{
		domain:     domain,
		serverName: serverName,
		serverType: serverType,
		services: make(map[string]string),
		vm:         lua4go.NewLuaVM(bind.NewDefault(), 1, 10, time.Second),
	}
}

func (s *scriptPlugin) Start() (svs []string, err error) {
	svs = make([]string,0 0)
	path := fmt.Sprintf("%s/%s/%s/script", s.domain, s.serverName, s.serverType)
	p, err := file.GetAbs(path)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return
	}
	for _, v := range files {
		if v.IsDir(){
			continue
		}		
		fileName:=v.Name()
		for _, method := range METHOD_NAME {
			if strings.HasPrefix(fileName, method+".") {
				filePath:=fmt.Sprintf("%s/%s",p,fileName)
				if err = s.vm.PreLoad(filePath); err != nil {
					return
				}
				svName:=fmt.Sprintf("%s.%s",fileName,method)
				s.services[svName]=filePath
				svs = append(svs,svName)
			}
		}
	}
	return
}

func (s *scriptPlugin) Close() error {
	s.vm.Close()
	return nil
}
func (s *scriptPlugin) Handle(name string, method string, service string, params string, c context.Context) (r *context.Response,err error) {
	svName:=fmt.Sprintf("%s.%s",service,method)
	f,ok:=range s.services[svName]
	 if !ok{
		return nil,fmt.Errorf("script plugin 未找到服务：",svName)
	 }
	 input:=lua4go.NewContext(params,c)
	result,m,err:=s.vm.Call(f,input)
	if err!=nil{
		return
	}
	r=&context.Response{Status:200,Params:m}
	return
}
