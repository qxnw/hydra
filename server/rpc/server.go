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

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/hydra/server/rpc/pb"
	"github.com/qxnw/lib4go/logger"
	"google.golang.org/grpc"
)

//RPCServer RPC Server
type RPCServer struct {
	domain     string
	server     *grpc.Server
	serverName string
	address    string
	process    *process
	ctxPool    sync.Pool
	ErrHandler Handler
	handlers   []Handler
	*serverOption
	port        int
	clusterPath string
	Router
	mu               sync.RWMutex
	apiRouters       []*rpcRouter
	localRRCServices []string
	remoteRPCService []string
	xsrf             *Auth
	jwt              *Auth
	basic            *Auth
	api              *Auth
}

//Version 获取当前版本号
func Version() string {
	return "0.0.1"
}

type serverOption struct {
	ip string
	*logger.Logger
	extHandlers  []Handler
	metric       *InfluxMetric
	limiter      *Limiter
	services     []string
	registry     server.IServiceRegistry
	registryRoot string
	running      bool
}

//Option 配置选项
type Option func(*serverOption)

//WithLogger 设置日志记录组件
func WithLogger(logger *logger.Logger) Option {
	return func(o *serverOption) {
		o.Logger = logger
	}
}

//WithInfluxMetric 设置基于influxdb的系统监控组件
func WithInfluxMetric(host string, dataBase string, userName string, password string, cron string) Option {
	return func(o *serverOption) {
		o.metric.RestartReport(host, dataBase, userName, password, cron, o.Logger)
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
		o.extHandlers = append(o.extHandlers, o.limiter)
	}
}

//WithRegistry 添加服务注册组件
func WithRegistry(r server.IServiceRegistry, root string) Option {
	return func(o *serverOption) {
		o.registry = r
		o.registryRoot = root
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
		o.extHandlers = append(o.extHandlers, handlers...)
	}
}

//NewRPCServer 初始化
func NewRPCServer(domain string, name string, services []string, opts ...Option) *RPCServer {
	s := &RPCServer{domain: domain, serverName: name, Router: NewRouter(), localRRCServices: services}
	s.serverOption = &serverOption{metric: NewInfluxMetric(), limiter: NewLimiter(map[string]int{})}
	s.process = newProcess(s)
	s.ErrHandler = Errors()

	for _, opt := range opts {
		opt(s.serverOption)
	}
	if s.Logger == nil {
		s.Logger = logger.GetSession(name, logger.CreateSession())
	}

	s.Use(Logging(),
		Recovery(false),
		s.metric,
		s.limiter,
		JWTFilter(),
		Return(),
		Param(),
		Contexts())
	s.Use(s.extHandlers...)
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
	s.Info("Listening on " + addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	s.server = grpc.NewServer()
	pb.RegisterRPCServer(s.server, s.process)

	startChan := make(chan error, 1)
	go func() {
		err = s.server.Serve(lis)
		startChan <- err
	}()
	select {
	case <-time.After(time.Second * 2):
		s.running = true
		err = s.registerService()
		return err
	case err := <-startChan:
		s.running = false
		return err
	}
}

//Close 关闭连接
func (s *RPCServer) Close() {
	s.running = false
	s.unRegisterService()
	if s.server != nil {
		s.Infof("rpc: Server closed(%s)", s.serverName)
		s.server.GracefulStop()
	}
}

//UpdateLimiter 更新限流规则
func (s *RPCServer) UpdateLimiter(limit map[string]int) {
	if s.limiter != nil {
		s.limiter.Update(limit)
	}
}

//SetInfluxMetric 重置metric
func (s *RPCServer) SetInfluxMetric(host string, dataBase string, userName string, password string, cron string) error {
	err := s.metric.RestartReport(host, dataBase, userName, password, cron, s.Logger)
	if err != nil {
		s.Error(err)
	}
	return err
}

//StopInfluxMetric stop metric
func (s *RPCServer) StopInfluxMetric() {
	s.metric.Stop()
}

//SetXSRF 设置XSRF安全认证参数
func (s *RPCServer) SetXSRF(enable bool, name string, secret string, exclude []string, expireAt int64) {
	name = fmt.Sprintf("__%s__", name)
	s.xsrf = &Auth{Enable: enable, Name: name, Secret: secret, Exclude: exclude, ExpireAt: expireAt}
}

//SetJWT 设置jwt安全认证参数
func (s *RPCServer) SetJWT(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	name = fmt.Sprintf("__%s__", name)
	s.jwt = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
}

//SetBasic 设置basic安全认证参数
func (s *RPCServer) SetBasic(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	name = fmt.Sprintf("__%s__", name)
	s.basic = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
}

//SetAPI 设置api安全认证参数
func (s *RPCServer) SetAPI(enable bool, name string, mode string, secret string, exclude []string, expireAt int64) {
	name = fmt.Sprintf("__%s__", name)
	s.api = &Auth{Enable: enable, Name: name, Secret: secret, Mode: mode, Exclude: exclude, ExpireAt: expireAt}
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
	s.apiRouters = routers
	for _, v := range routers {
		s.Route(v.Method, v.Path, v.Handler, v.Middlewares...)
	}
}

//Request 设置Request路由
func (s *RPCServer) Request(service string, c interface{}, middlewares ...Handler) {
	s.Route([]string{"REQUEST"}, service, c, middlewares...)
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
}
