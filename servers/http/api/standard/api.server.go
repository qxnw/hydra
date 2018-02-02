package standard

import (
	"context"
	"fmt"
	x "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/logger"
)

//Server api服务器
type Server struct {
	*option
	conf    *conf.ApiServerConf
	engine  *x.Server
	running bool
	proto   string
	port    int
}

//New 创建api服务器
func New(conf *conf.ApiServerConf, routers []*conf.Router, opts ...Option) (t *Server, err error) {
	t = &Server{conf: conf}
	t.option = &option{metric: middleware.NewMetric(t.conf.ServerConf), static: &middleware.StaticOptions{Enable: false}}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(conf.GetFullName(), logger.CreateSession())
	}
	t.engine = &x.Server{
		ReadHeaderTimeout: time.Second * 3,
		ReadTimeout:       time.Second * 3,
		WriteTimeout:      time.Second * 3,
		MaxHeaderBytes:    1 << 20,
	}
	if routers != nil {
		t.engine.Handler, err = t.getHandler(routers)
	}
	return
}

// Run the http server
func (s *Server) Run(address ...interface{}) error {
	addr := s.getAddress(address...)
	s.proto = "http"
	s.engine.Addr = addr
	s.running = true
	errChan := make(chan error, 1)
	go func(ch chan error) {
		if err := s.engine.ListenAndServe(); err != nil {
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

//RunTLS RunTLS server
func (s *Server) RunTLS(certFile, keyFile string, address ...interface{}) error {
	addr := s.getAddress(address...)
	s.proto = "https"
	s.engine.Addr = addr
	s.running = true
	errChan := make(chan error, 1)
	go func(ch chan error) {
		if err := s.engine.ListenAndServeTLS(certFile, keyFile); err != nil {
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
		ctx, cannel := context.WithTimeout(context.Background(), timeout)
		defer cannel()
		if err := s.engine.Shutdown(ctx); err != nil {
			if err == x.ErrServerClosed {
				s.Errorf("%s:已关闭", s.conf.GetFullName())
				return
			}
			s.Errorf("%s关闭出现错误:%v", s.conf.GetFullName(), err)
		}
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
