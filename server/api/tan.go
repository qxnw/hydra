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

//WebServer web服务器
type WebServer struct {
	server     *http.Server
	serverName string
	proto      string
	port       int
	Router
	handlers    []Handler
	ErrHandler  Handler
	ctxPool     sync.Pool
	respPool    sync.Pool
	clusterPath string
	mu          sync.RWMutex
	*webServerOption
	running bool
}

var (
	//ClassicHandlers 标准插件
	ClassicHandlers = []Handler{
		Logging(),
		Recovery(false),
		Compresses([]string{}),
		Static(StaticOptions{Prefix: "public"}),
		Return(),
		Param(),
		Contexts(),
	}
)

type webServerOption struct {
	ip           string
	logger       *logger.Logger
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
		o.logger = logger
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
		o.metric.RestartReport(host, dataBase, userName, password, timeSpan)
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

//Logger 获取日志组件
func (t *WebServer) Logger() *logger.Logger {
	return t.logger
}

//SetName 设置组件的server name
func (t *WebServer) SetName(name string) {
	t.serverName = name
}

//SetHost 设置组件的host name
func (t *WebServer) SetHost(host string) {
	if len(host) > 0 {
		t.hostNames = strings.Split(host, ",")
	}

}

//SetInfluxMetric 重置metric
func (t *WebServer) SetInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) {
	err := t.metric.RestartReport(host, dataBase, userName, password, timeSpan)
	if err != nil {
		t.logger.Error("启动metric失败：", err)
	}
}

//StopInfluxMetric stop metric
func (t *WebServer) StopInfluxMetric() {
	t.metric.Stop()
}

//SetRouters 设置路由规则
func (t *WebServer) SetRouters(routers ...*webRouter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Router = NewRouter()
	for _, v := range routers {
		t.Route(v.Method, v.Path, v.Handler, v.Middlewares...)
	}
}

// Run the http server. Listening on os.GetEnv("PORT") or 8000 by default.
func (t *WebServer) Run(address ...interface{}) error {
	addr := t.getAddress(address...)
	t.logger.Info("Listening on http://" + addr)
	t.proto = "http"
	t.server = &http.Server{Addr: addr, Handler: t}
	err := t.registryServer()
	if err != nil {
		t.logger.Error(err)
		return err
	}
	t.running = true
	err = t.server.ListenAndServe()
	if err != nil {
		t.running = false
		t.logger.Error(err)
		return err
	}

	return nil
}

//RunTLS RunTLS server
func (t *WebServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	addr := t.getAddress(address...)
	t.logger.Info("Listening on https://" + addr)
	t.proto = "https"
	t.server = &http.Server{Addr: addr, Handler: t}
	err := t.registryServer()
	if err != nil {
		t.logger.Error(err)
		return err
	}
	t.running = true
	err = t.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		t.running = false
		t.logger.Error(err)
		return err
	}
	return nil
}

//New create new server
func New(name string, opts ...Option) *WebServer {
	t := &WebServer{
		serverName:      name,
		Router:          NewRouter(),
		ErrHandler:      Errors(),
		webServerOption: &webServerOption{host: Host(), metric: NewInfluxMetric(), logger: logger.GetSession(name, logger.CreateSession())},
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
		Static(StaticOptions{Prefix: "public"}),
		Return(),
		Param(),
		Contexts())
	handlers = append(handlers, t.webServerOption.handlers...)
	//构建缓存
	t.ctxPool.New = func() interface{} {
		return &Context{
			tan: t,
		}
	}

	t.respPool.New = func() interface{} {
		return &responseWriter{}
	}

	t.Use(handlers...)
	return t
}

//Shutdown shutdown server
func (t *WebServer) Shutdown(timeout time.Duration) {
	t.running = false
	t.unregistryServer()
	if t.server != nil {
		xt, _ := ctx.WithTimeout(ctx.Background(), timeout)
		t.server.Shutdown(xt)
	}

}

//GetAddress 获取当前服务地址
func (t *WebServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", t.proto, t.ip, t.port)
}