// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	ctx "context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/logger"
)

//Version 系统版本号
func Version() string {
	return "0.0.0.1"
}

//HTTPServer http服务器
type HTTPServer struct {
	domain     string
	server     *http.Server
	serverName string
	proto      string
	port       int
	typeName   string
	Router
	handlers    []Handler
	ErrHandler  Handler
	ctxPool     sync.Pool
	respPool    sync.Pool
	clusterPath string
	mu          sync.RWMutex
	*webServerOption
	Running              bool
	Headers              map[string]string
	xsrf                 *Auth
	jwt                  *Auth
	basic                *Auth
	api                  *Auth
	onlyAllowAjaxRequest bool
}

type webServerOption struct {
	ip string
	*logger.Logger
	registry     server.IServiceRegistry
	metric       *InfluxMetric
	hostNames    []string
	host         Handler
	handlers     []Handler
	registryRoot string
}

//Option 配置选项
type Option func(*webServerOption)

//WithLogger 设置日志记录组件
func WithLogger(logger *logger.Logger) Option {
	return func(o *webServerOption) {
		o.Logger = logger
	}
}

//WithIP 设置ip地址
func WithIP(ip string) Option {
	return func(o *webServerOption) {
		o.ip = ip
	}
}

//WithInfluxMetric 设置基于influxdb的系统监控组件
func WithInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) Option {
	return func(o *webServerOption) {
		o.metric.RestartReport(host, dataBase, userName, password, timeSpan, o.Logger)
	}
}

//WithHost 添加插件
func WithHost(host string) Option {
	return func(o *webServerOption) {
		o.hostNames = strings.Split(host, ",")
	}
}

//WithRegistry 添加服务注册组件
func WithRegistry(r server.IServiceRegistry, root string) Option {
	return func(o *webServerOption) {
		o.registry = r
		o.registryRoot = root
	}
}

//WithHandlers 添加插件
func WithHandlers(handlers ...Handler) Option {
	return func(o *webServerOption) {
		o.handlers = handlers
	}
}

//GetLogger 获取日志组件
func (t *HTTPServer) GetLogger() *logger.Logger {
	return t.Logger
}

//SetName 设置组件的server name
func (t *HTTPServer) SetName(name string) {
	t.serverName = name
}

//OnlyAllowAjaxRequest 只允许ajax请求
func (t *HTTPServer) OnlyAllowAjaxRequest(allow bool) {
	t.onlyAllowAjaxRequest = allow
}

//SetHost 设置组件的host name
func (t *HTTPServer) SetHost(host string) {
	if len(host) > 0 {
		t.hostNames = strings.Split(host, ",")
	}

}

//SetInfluxMetric 重置metric
func (t *HTTPServer) SetInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) {
	err := t.metric.RestartReport(host, dataBase, userName, password, timeSpan, t.Logger)
	if err != nil {
		t.Error("启动metric失败：", err)
	}
}

//StopInfluxMetric stop metric
func (t *HTTPServer) StopInfluxMetric() {
	t.metric.Stop()
}

//SetRouters 设置路由规则
func (t *HTTPServer) SetRouters(routers ...*WebRouter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Router = NewRouter()
	for _, v := range routers {
		t.Route(v.Method, v.Path, v.Handler, v.Middlewares...)
	}
}

//SetXSRF 设置XSRF安全认证参数
func (t *HTTPServer) SetXSRF(enable bool, name string, secret string, exclude []string, expireAt int64) {
	t.xsrf = &Auth{Enable: enable, Name: name, Secret: secret, Exclude: exclude, ExpireAt: expireAt}
}

//SetJWT 设置jwt安全认证参数
func (t *HTTPServer) SetJWT(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	t.jwt = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
}

//SetBasic 设置basic安全认证参数
func (t *HTTPServer) SetBasic(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	t.basic = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
}

//SetAPI 设置api安全认证参数
func (t *HTTPServer) SetAPI(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	t.api = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
}

// Run the http server. Listening on os.GetEnv("PORT") or 8000 by default.
func (t *HTTPServer) Run(address ...interface{}) error {
	addr := t.getAddress(address...)
	t.Info("Listening on http://" + addr)
	t.proto = "http"
	t.server = &http.Server{Addr: addr, Handler: t}
	err := t.registryServer()
	if err != nil {
		return err
	}
	t.Running = true
	err = t.server.ListenAndServe()
	if err != nil {
		t.Running = false
		return err
	}

	return nil
}

//RunTLS RunTLS server
func (t *HTTPServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	addr := t.getAddress(address...)
	t.Info("Listening on https://" + addr)
	t.proto = "https"
	t.server = &http.Server{Addr: addr, Handler: t}
	err := t.registryServer()
	if err != nil {
		t.Error(err)
		return err
	}
	t.Running = true
	err = t.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		t.Running = false
		t.Error(err)
		return err
	}
	return nil
}

//NewAPI create new server
func NewAPI(domain string, name string, opts ...Option) *HTTPServer {
	handlers := make([]Handler, 0, 4)
	handlers = append(handlers,
		APIReturn(),
		Param(),
		Contexts())
	opts = append(opts, WithHandlers(handlers...))
	return New(domain, name, "api", opts...)
}

//New create new server
func New(domain string, name string, typeName string, opts ...Option) *HTTPServer {
	t := &HTTPServer{
		typeName:        typeName,
		domain:          domain,
		serverName:      name,
		Router:          NewRouter(),
		ErrHandler:      Errors(),
		webServerOption: &webServerOption{host: Host(), metric: NewInfluxMetric(), Logger: logger.GetSession(name, logger.CreateSession())},
	}

	//转换配置项
	for _, opt := range opts {
		opt(t.webServerOption)
	}
	handlers := make([]Handler, 0, 8)
	handlers = append(handlers,
		Logging(),
		Recovery(false),
		t.webServerOption.host,
		t.webServerOption.metric,
		Compresses([]string{}),
		WriteHeader(),
		Static(StaticOptions{Prefix: ""}),
		OnlyAllowAjaxRequest(),
		XSRFFilter(),
		JWTFilter(),
	)
	handlers = append(handlers, t.webServerOption.handlers...)
	//构建缓存
	t.ctxPool.New = func() interface{} {
		return &Context{
			Server: t,
		}
	}

	t.respPool.New = func() interface{} {
		return &responseWriter{}
	}

	t.Use(handlers...)
	return t
}

//SetStatic 设置静态文件路由
func (t *HTTPServer) SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) {
	t.handlers[6] = Static(StaticOptions{
		Enable:     enable,
		Prefix:     prefix,
		RootPath:   dir,
		FilterExts: exts,
	})
}

//Shutdown shutdown server
func (t *HTTPServer) Shutdown(timeout time.Duration) {
	t.Running = false
	t.unregistryServer()
	if t.server != nil {
		xt, _ := ctx.WithTimeout(ctx.Background(), timeout)
		t.server.Shutdown(xt)
	}
}

//GetAddress 获取当前服务地址
func (t *HTTPServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", t.proto, t.ip, t.port)
}
func (t *HTTPServer) SetHeader(headers map[string]string) {
	t.Headers = headers
}
