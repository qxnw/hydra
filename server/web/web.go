package web

import (
	"os"
	"path/filepath"

	"strings"

	"fmt"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/hydra/server/web/template"
)

//WebServer web server
type WebServer struct {
	*api.HTTPServer
	viewRoot   string
	viewExt    string
	viewTmpl   *template.Set
	leftDelim  string
	rightDelim string
	views      []string
	errorView  string
}

//New 构建WEB服务器
func New(domain string, name string, opts ...api.Option) *WebServer {
	server := &WebServer{viewTmpl: &template.Set{}, viewRoot: "../views", viewExt: ".html", leftDelim: "{{", rightDelim: "}}", errorView: "error"}
	server.views = make([]string, 0, 20)
	handlers := make([]api.Handler, 0, 4)
	handlers = append(handlers,
		server.Return(),
		api.Param(),
		api.Contexts())
	opts = append(opts, api.WithHandlers(handlers...))
	server.HTTPServer = api.New(domain, name, "web", opts...)
	return server
}

//SetViewsPath 设置view根路径
func (w *WebServer) SetViewsPath(path string) {
	w.viewRoot = path
}

//SetDelims 设置绑定表达式的
func (w *WebServer) SetDelims(left string, right string) {
	w.leftDelim = left
	w.rightDelim = right
}

//loadTmpl 加载模板
func (w *WebServer) loadTmpl() error {
	w.viewTmpl.Delims(w.leftDelim, w.rightDelim)
	err := filepath.Walk(w.viewRoot, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
			return nil
		}
		if !strings.HasSuffix(path, w.viewExt) {
			return nil
		}
		_, er := w.viewTmpl.ParseFiles(path)
		if er != nil {
			er = fmt.Errorf("转换模板文件失败:%s(err:%v)", path, er)
			return er
		}
		w.views = append(w.views, path)
		return nil
	})
	if err != nil {
		err = fmt.Errorf("加载模板文件失败：%s,%v", w.viewRoot, err)
		return err
	}
	return nil
}

//Run 启动服务器
func (w *WebServer) Run(address ...interface{}) error {
	if err := w.loadTmpl(); err != nil {
		return err
	}
	return w.HTTPServer.Run(address...)
}

//RunTLS 启动服务器
func (w *WebServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	if err := w.loadTmpl(); err != nil {
		return err
	}
	return w.HTTPServer.RunTLS(certFile, keyFile, address...)
}

//ExistView 是否存在view
func (w *WebServer) ExistView(path string) bool {
	for _, v := range w.views {
		if v == path {
			return true
		}
	}
	return false
}
