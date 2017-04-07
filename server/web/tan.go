// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	ctx "context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/qxnw/hydra/context"
)

//Version 系统版本号
func Version() string {
	return "0.5.2.1214"
}

//WebServer web服务器
type WebServer struct {
	server     *http.Server
	serverName string
	proto      string
	port       int
	Router
	handlers   []Handler
	ErrHandler Handler
	ctxPool    sync.Pool
	respPool   sync.Pool
	mu         sync.RWMutex
	*webServerOption
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
	ip        string
	logger    Logger
	register  context.IServiceRegistry
	metric    *InfluxMetric
	hostNames []string
	host      Handler
	handlers  []Handler
}

//Option 配置选项
type Option func(*webServerOption)

//WithLogger 设置日志记录组件
func WithLogger(logger Logger) Option {
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
		if o.metric == nil {
			o.metric = NewInfluxMetric()
		}
		o.metric.RestartReport(host, dataBase, userName, password, timeSpan)
	}
}

//WithHost 添加插件
func WithHost(host string) Option {
	return func(o *webServerOption) {
		if o.host == nil {
			o.host = Host()
		}
		o.hostNames = strings.Split(host, ",")

	}
}

//WithHandlers 添加插件
func WithHandlers(handlers ...Handler) Option {
	return func(o *webServerOption) {
		o.handlers = handlers
	}
}

//Logger 获取日志组件
func (t *WebServer) Logger() Logger {
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
	t.metric.RestartReport(host, dataBase, userName, password, timeSpan)
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
	err := t.server.ListenAndServe()
	if err != nil {
		t.logger.Error(err)
	}
	t.registerService()
	return err
}

//RunTLS RunTLS server
func (t *WebServer) RunTLS(certFile, keyFile string, address ...interface{}) error {
	addr := t.getAddress(address...)
	t.logger.Info("Listening on https://" + addr)
	t.proto = "https"
	t.server = &http.Server{Addr: addr, Handler: t}
	err := t.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		t.logger.Error(err)
	}
	t.registerService()
	return err
}

//New create new server
func New(name string, opts ...Option) *WebServer {
	t := &WebServer{
		serverName:      name,
		Router:          NewRouter(),
		handlers:        make([]Handler, 0),
		ErrHandler:      Errors(),
		webServerOption: &webServerOption{},
	}
	//转换配置项
	for _, opt := range opts {
		opt(t.webServerOption)
	}
	//设置日志
	if t.webServerOption.logger == nil {
		t.webServerOption.logger = NewLogger(name, os.Stdout)
	}

	//处理外部插件
	if t.webServerOption.host == nil {
		t.webServerOption.host = Host()
	}
	if t.webServerOption.metric == nil {
		t.webServerOption.metric = NewInfluxMetric()
	}

	handlers := make([]Handler, 0, len(t.webServerOption.handlers)+len(ClassicHandlers)+2)
	handlers = append(handlers, t.webServerOption.host)
	handlers = append(handlers, t.webServerOption.metric)
	if len(t.webServerOption.handlers) > 0 {
		handlers = append(handlers, t.webServerOption.handlers...)
	}
	handlers = append(handlers, ClassicHandlers...)
	//构建缓存
	t.ctxPool.New = func() interface{} {
		return &Context{
			tan:    t,
			Logger: t.webServerOption.logger,
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
	t.unRegisterService()
	if t.server != nil {
		xt, _ := ctx.WithTimeout(ctx.Background(), timeout)
		t.server.Shutdown(xt)
	}

}

//GetAddress 获取当前服务地址
func (t *WebServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", t.proto, t.ip, t.port)
}
