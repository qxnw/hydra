package web

import (
	"fmt"
	"os"
	"path/filepath"

	"strings"

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
}

//New 构建WEB服务器
func New(domain string, name string, opts ...api.Option) *WebServer {
	server := &WebServer{viewTmpl: &template.Set{}, viewRoot: "../views", viewExt: ".html", leftDelim: "{{", rightDelim: "}}"}
	handlers := make([]api.Handler, 0, 8)
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
func (w *WebServer) SetDelims(left string, right string) {
	w.leftDelim = left
	w.rightDelim = right
}

//Run 启动服务器
func (w *WebServer) Run(address ...interface{}) error {
	//w.viewTmpl.Delims(w.leftDelim, w.rightDelim)
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
		return er
	})
	if err != nil {
		return err
	}

	return w.HTTPServer.Run(address...)
}

//RunTLS 启动服务器
func (w *WebServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	set := w.viewTmpl.Delims(w.leftDelim, w.rightDelim)
	_, err := set.ParseGlob(fmt.Sprintf("%s/*.%s", w.viewRoot, w.viewExt))
	if err != nil {
		return err
	}
	return w.HTTPServer.RunTLS(certFile, keyFile, address...)
}
