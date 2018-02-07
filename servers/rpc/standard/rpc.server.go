package standard

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/qxnw/hydra/servers/rpc/standard/pb"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
	"google.golang.org/grpc"
)

//Server api服务器
type Server struct {
	*option
	conf   *conf.RpcServerConf
	engine *grpc.Server
	*Processor
	running bool
	proto   string
	port    int
	addr    string
}

//New 创建api服务器
func New(conf *conf.RpcServerConf, routers []*conf.Router, opts ...Option) (t *Server, err error) {
	t = &Server{conf: conf}
	t.option = &option{metric: middleware.NewMetric(t.conf.ServerConf)}
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
	} else {
		t.Processor = NewProcessor()
	}

	return
}

// Run the http server
func (s *Server) Run(address ...interface{}) error {
	pb.RegisterRPCServer(s.engine, s.Processor)
	addr := s.getAddress(address...)
	s.proto = "tcp"
	s.addr = addr
	s.running = true
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
		s.running = false
		return err
	}
}

//Shutdown 关闭服务器
func (s *Server) Shutdown(timeout time.Duration) {
	if s.engine != nil {
		s.running = false
		s.engine.GracefulStop()
		time.Sleep(time.Second)
		s.Infof("%s:已关闭", s.conf.GetFullName())

	}
}

//GetAddress 获取当前服务地址
func (s *Server) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", s.proto, s.ip, s.port)
}

//GetStatus 获取当前服务器状态
func (s *Server) GetStatus() string {
	if s.running {
		return servers.ST_RUNNING
	}
	return servers.ST_STOP
}

func (s *Server) getAddress(args ...interface{}) string {
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
