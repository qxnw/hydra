package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
)

//CoreServer core server服务器
type CoreServer struct {
	*option
	conf    *conf.ServerConf
	Running bool
	Proto   string
	Port    int
	Addr    string
}

//New 创建api服务器
func New(conf *conf.ServerConf, opts ...Option) (t *CoreServer) {
	t = &CoreServer{conf: conf}
	t.option = &option{Metric: middleware.NewMetric(t.conf)}
	for _, opt := range opts {
		opt(t.option)
	}
	if t.Logger == nil {
		t.Logger = logger.GetSession(conf.GetFullName(), logger.CreateSession())
	}
	return
}

//GetAddress 获取当前服务地址
func (s *CoreServer) GetAddress() string {
	return fmt.Sprintf("%s://%s:%d", s.Proto, s.conf.IP, s.Port)
}

//GetStatus 获取当前服务器状态
func (s *CoreServer) GetStatus() string {
	if s.Running {
		return servers.ST_RUNNING
	}
	return servers.ST_STOP
}

func (s *CoreServer) GetAvailableAddress(args ...interface{}) string {
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
	s.Port = port
	addr := host + ":" + strconv.FormatInt(int64(port), 10)
	return addr
}
