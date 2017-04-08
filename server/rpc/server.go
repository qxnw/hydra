package rpc

import (
	"fmt"
	"net"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"os"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server/rpc/pb"
	"google.golang.org/grpc"
)

//RPCServer RPC Server
type RPCServer struct {
	server     *grpc.Server
	serverName string
	address    string
	process    *process
	ctxPool    sync.Pool
	ErrHandler Handler
	*serverOption
	port int

	Router
	mu sync.RWMutex
}

//Version 获取当前版本号
func Version() string {
	return "0.0.1"
}

type serverOption struct {
	ip       string
	logger   Logger
	handlers []Handler
	metric   *InfluxMetric
	limiter  *Limiter
	services []string
	registry context.IServiceRegistry
}

//Option 配置选项
type Option func(*serverOption)

//WithLogger 设置日志记录组件
func WithLogger(logger Logger) Option {
	return func(o *serverOption) {
		o.logger = logger
	}
}

//WithInfluxMetric 设置基于influxdb的系统监控组件
func WithInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) Option {
	return func(o *serverOption) {
		o.metric.RestartReport(host, dataBase, userName, password, timeSpan)
	}
}

//WithIP 设置ip地址
func WithIP(ip string) Option {
	return func(o *serverOption) {
		o.ip = ip
	}
}

//WithLimiter 设置流量限制组件
func WithLimiter(limit map[string]int) Option {
	return func(o *serverOption) {
		o.handlers = append(o.handlers, o.limiter)
	}
}

//WithRegister 设置服务注册组件
func WithRegister(i context.IServiceRegistry) Option {
	return func(o *serverOption) {
		o.registry = i
	}
}

//WithServices 设置服务注册组件
func WithServices(services ...string) Option {
	return func(o *serverOption) {
		o.services = services
	}
}

//WithPlugins 添加插件
func WithPlugins(handlers ...Handler) Option {
	return func(o *serverOption) {
		o.handlers = append(o.handlers, handlers...)
	}
}

var (
	//ClassicHandlers 标准插件
	ClassicHandlers = []Handler{
		Logging(),
		Recovery(false),
		Return(),
		Param(),
		Contexts(),
	}
)

//NewRPCServer 初始化
func NewRPCServer(name string, opts ...Option) *RPCServer {
	s := &RPCServer{serverName: name, Router: NewRouter()}
	s.serverOption = &serverOption{logger: NewLogger(name, os.Stdout), metric: NewInfluxMetric(),limiter:NewLimiter(map[string]int{}}

	s.process = &process{srv: s}
	s.ErrHandler = Errors()

	for _, opt := range opts {
		opt(s.serverOption)
	}
	s.Use(Logging(),
		Recovery(false),
		s.metric,
		s.limiter,
		s.handlers...,
		Return(),
		Param(),
		Contexts())
	return s
}

//Use 添加全局插件
func (s *RPCServer) Use(handlers ...Handler) {
	s.handlers = append(s.handlers, handlers...)
}

//Run 运行服务堵塞当前线程直到系统被中断退出
func (s *RPCServer) Run(address string) (err error) {
	err = s.Start(address)
	if err != nil {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	<-ch
	s.Close()
	return nil
}

//Start 启动RPC服务器
func (s *RPCServer) Start(address string) (err error) {
	addr := s.getAddress(address)
	s.logger.Info("Listening on " + addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	s.server = grpc.NewServer()
	pb.RegisterRPCServer(s.server, s.process)
	go func() {
		err = s.server.Serve(lis)

	}()
	if err != nil {
		return
	}
	s.registerService()
	return
}

//Close 关闭连接
func (s *RPCServer) Close() {
	s.unRegisterService()
	if s.server != nil {
		s.logger.Error("rpc: Server closed")
		s.server.GracefulStop()
	}
}

//Logger 获取日志组件
func (s *RPCServer) Logger() Logger {
	return s.logger
}

//UpdateLimiter 更新限流规则
func (s *RPCServer) UpdateLimiter(limit map[string]int) {
	if s.limiter != nil {
		s.limiter.Update(limit)
	}
}

//SetInfluxMetric 重置metric
func (s *RPCServer) SetInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) {
	s.metric.RestartReport(host, dataBase, userName, password, timeSpan)
}

//StopInfluxMetric stop metric
func (s *RPCServer) StopInfluxMetric() {
	s.metric.Stop()
}

//SetName 设置组件的server name
func (s *RPCServer) SetName(name string) {
	s.serverName = name
}

//SetRouters 设置路由规则
func (s *RPCServer) SetRouters(routers ...*rpcRouter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Router = NewRouter()
	for _, v := range routers {
		s.Route(v.Method, v.Path, v.Handler, v.Middlewares...)
	}
}

//Request 设置Request路由
func (s *RPCServer) Request(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"REQUEST"}, service, c, middlewares...)
}

//Query 设置Query路由
func (s *RPCServer) Query(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"QUERY"}, service, c, middlewares...)
}

//Insert 设置Insert路由
func (s *RPCServer) Insert(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"INSERT"}, service, c, middlewares...)
}

//Delete 设置Delete路由
func (s *RPCServer) Delete(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"DELETE"}, service, c, middlewares...)
}

//Update 设置Update路由
func (s *RPCServer) Update(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"UPDATE"}, service, c, middlewares...)
}
func (s *RPCServer) getAddress(args ...interface{}) string {
	var host string
	var port int

	if len(args) == 1 {
		switch arg := args[0].(type) {
		case string:
			addrs := strings.Split(args[0].(string), ":")
			if len(addrs) == 1 {
				host = addrs[0]
			} else if len(addrs) >= 2 {
				host = addrs[0]
				_port, _ := strconv.ParseInt(addrs[1], 10, 0)
				port = int(_port)
			}
		case int:
			port = arg
		}
	} else if len(args) >= 2 {
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = arg
		}
	}

	if len(host) == 0 {
		host = s.ip
		if host == "" {
			host = "0.0.0.0"
		}

	}
	if port == 0 {
		port = 8000
	}
	s.port = port
	return host + ":" + strconv.FormatInt(int64(port), 10)
}

//GetAddress 获取当前服务地址
func (s *RPCServer) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.ip, s.port)
	//return fmt.Sprintf("tcp://%s:%d", s.ip, s.port)
}
