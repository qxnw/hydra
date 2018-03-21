package http

import (
	"context"
	"fmt"
	x "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/lib4go/logger"
)

//ApiServer api服务器
type ApiServer struct {
	*option
	conf    *conf.MetadataConf
	engine  *x.Server
	running string
	proto   string
	host    string
	port    int
}

//NewApiServer 创建api服务器
func NewApiServer(name string, addr string, routers []*conf.Router, opts ...Option) (t *ApiServer, err error) {
	t = &ApiServer{conf: &conf.MetadataConf{
		Name: name,
		Type: "api",
	}}
	t.option = &option{
		metric:            middleware.NewMetric(t.conf),
		readHeaderTimeout: 6,
		readTimeout:       6,
		writeTimeout:      6}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(name, logger.CreateSession())
	}
	t.engine = &x.Server{
		Addr:              t.getAddress(addr),
		ReadHeaderTimeout: time.Second * time.Duration(t.option.readHeaderTimeout),
		ReadTimeout:       time.Second * time.Duration(t.option.readTimeout),
		WriteTimeout:      time.Second * time.Duration(t.option.writeTimeout),
		MaxHeaderBytes:    1 << 20,
	}
	if routers != nil {
		t.engine.Handler, err = t.getHandler(routers)
	}
	return
}

// Run the http server
func (s *ApiServer) Run() error {
	s.proto = "http"
	s.running = servers.ST_RUNNING
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
		s.running = servers.ST_STOP
		return err
	}
}

//RunTLS RunTLS server
func (s *ApiServer) RunTLS(certFile, keyFile string) error {
	s.proto = "https"
	s.running = servers.ST_RUNNING
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
		s.running = servers.ST_STOP
		return err
	}
}

//Shutdown 关闭服务器
func (s *ApiServer) Shutdown(timeout time.Duration) {
	if s.engine != nil {
		s.metric.Stop()
		s.running = servers.ST_STOP
		ctx, cannel := context.WithTimeout(context.Background(), timeout)
		defer cannel()
		if err := s.engine.Shutdown(ctx); err != nil {
			if err == x.ErrServerClosed {
				s.Infof("%s:已关闭", s.conf.Name)
				return
			}
			s.Errorf("关闭出现错误:%v", err)
		}
	}
}

//GetAddress 获取当前服务地址
func (s *ApiServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", s.proto, s.host, s.port)
}

//GetStatus 获取当前服务器状态
func (s *ApiServer) GetStatus() string {
	return s.running
}

func (s *ApiServer) getAddress(args ...interface{}) string {
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
	switch host {
	case s.ip, "0.0.0.0":
		s.host = s.ip
	default:
		s.host = host
	}
	addr := host + ":" + fmt.Sprint(s.port)
	return addr
}
