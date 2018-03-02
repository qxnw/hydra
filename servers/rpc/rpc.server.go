package rpc

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/rpc/pb"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
	"google.golang.org/grpc"
)

//RpcServer rpc服务器
type RpcServer struct {
	*option
	conf   *conf.ServerConf
	engine *grpc.Server
	*Processor
	running string
	proto   string
	port    int
	addr    string
}

//NewRpcServer 创建rpc服务器
func NewRpcServer(conf *conf.ServerConf, routers []*conf.Router, opts ...Option) (t *RpcServer, err error) {
	t = &RpcServer{conf: conf}
	t.option = &option{metric: middleware.NewMetric(t.conf)}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(conf.GetFullName(), logger.CreateSession())
	}
	t.engine = grpc.NewServer()
	if routers != nil {
		t.Processor, err = t.getProcessor(routers)
		if err != nil {
			return
		}
	}
	//else {
	//t.Processor = NewProcessor()
	//}

	return
}

// Run the http server
func (s *RpcServer) Run(address ...interface{}) error {
	pb.RegisterRPCServer(s.engine, s.Processor)
	addr := s.getAddress(address...)
	s.proto = "tcp"
	s.addr = addr
	s.running = servers.ST_RUNNING
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			ch <- err
			return
		}
		if err := s.engine.Serve(lis); err != nil {
			ch <- err
		}
	}(errChan)
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		s.running = servers.ST_STOP
		return err
	}
}

//Shutdown 关闭服务器
func (s *RpcServer) Shutdown(timeout time.Duration) {
	if s.engine != nil {
		s.running = servers.ST_STOP
		s.engine.GracefulStop()
		time.Sleep(time.Second)
		s.Infof("%s:已关闭", s.conf.GetFullName())

	}
}

//GetAddress 获取当前服务地址
func (s *RpcServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", s.proto, s.ip, s.port)
}

//GetStatus 获取当前服务器状态
func (s *RpcServer) GetStatus() string {
	return s.running
}

func (s *RpcServer) getAddress(args ...interface{}) string {
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
	addr := host + ":" + strconv.FormatInt(int64(port), 10)
	return addr
}
