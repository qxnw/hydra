package web

import (
	"strings"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/hydra/server/web/template"
)

//WebServer web server
type WebServer struct {
	*api.HTTPServer
	views      []string
	tmpl       []string
	viewTmpl   *template.Set
	leftDelim  string
	rightDelim string
}

//New 构建WEB服务器
func New(domain string, name string, opts ...api.Option) *WebServer {
	handlers := make([]api.Handler, 0, 8)
	handlers = append(handlers,
		Return(),
		api.Param(),
		api.Contexts())
	opts = append(opts, api.WithHandlers(handlers...))
	server := &WebServer{viewTmpl: &template.Set{}, views: []string{"../views"}, tmpl: []string{".html", "htm", "tmpl"}}
	server.HTTPServer = api.New(domain, name, "web", opts...)
	return server
}

//SetViewsPath 设置view根路径
func (w *WebServer) SetViewsPath(path string) {
	w.views = strings.Split(path, ";")
}
func (w *WebServer) SetDelims(left string, right string) {
	w.leftDelim = left
	w.rightDelim = right
}

//Run 启动服务器
func (w *WebServer) Run(address ...interface{}) error {
	//set := w.viewTmpl.Delims(w.leftDelim, w.rightDelim)
	//for _, v := range w.views {
	//set.ParseGlob(fmt.Sprintf("%s/%s.%s", v))
	//}
	return w.HTTPServer.Run(address...)
}

//RunTLS 启动服务器
func (w *WebServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	return w.HTTPServer.RunTLS(certFile, keyFile, address...)
}
