package standard

import (
	"net"
	"time"

	"github.com/qxnw/hydra/servers/rpc/standard/pb"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/core"
	"google.golang.org/grpc"
)

//Server api服务器
type Server struct {
	conf   *conf.RpcServerConf
	engine *grpc.Server
	*Processor
	*core.CoreServer
}

//New 创建api服务器
func New(conf *conf.RpcServerConf, routers []*conf.Router, opts ...core.Option) (t *Server, err error) {
	t = &Server{conf: conf}
	t.CoreServer = core.New(conf.ServerConf, opts...)
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
	addr := s.GetAvailableAddress(address...)
	s.Info("--------启动rpc:", addr)
	s.Proto = "tcp"
	s.Addr = addr
	s.Running = true
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
		s.Running = false
		return err
	}
}

//Shutdown 关闭服务器
func (s *Server) Shutdown(timeout time.Duration) {
	if s.engine != nil {
		s.Running = false
		s.engine.GracefulStop()
		time.Sleep(time.Second)
		s.Errorf("%s:已关闭", s.conf.GetFullName())
	}
}
