package logger

import (
	"fmt"
	"sync"

	"time"

	"github.com/golang/snappy"
	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

type rpcLoggerSetting struct {
	level    string
	layout   string
	domain   string
	server   string
	service  string
	interval string
}

type rpcWriter struct {
	cdomain    string
	address    string
	rpcInvoker *rpc.Invoker
	*rpcLoggerSetting
	logger     *logger.Logger
	writeError bool
	closeChan  chan struct{}
	once       sync.Once
}

func newRPCWriter(domain string, address string, logger *logger.Logger) (r *rpcWriter, err error) {
	r = &rpcWriter{
		closeChan:        make(chan struct{}),
		cdomain:          domain,
		address:          address,
		rpcLoggerSetting: &rpcLoggerSetting{},
		logger:           logger,
	}
	registry, err := registry.NewRegistryWithAddress(address, logger)
	if err != nil {
		err = fmt.Errorf("初始化注册中心失败：%s:%v", address, err)
		return nil, err
	}
	path := fmt.Sprintf("/%s/var/global/logger", domain)
	buff, err := r.getConfig(registry, path)
	if err != nil {
		return nil, err
	}
	loggerConf, err := conf.NewJSONConfWithJson(string(buff), 0, nil)
	if err != nil {
		err = fmt.Errorf("rpc日志配置错误:%s,%v", string(buff), err)
		return
	}
	if _, err = loggerConf.GetSectionString("layout"); err != nil {
		err = fmt.Errorf("rpc日志的interval字段配置有误:%v", string(buff))
		return
	}
	if loggerConf.String("server") == "" || loggerConf.String("service") == "" {
		err = fmt.Errorf("rpc日志配置字段server,service不能为空:%v", string(buff))
		return
	}
	r.interval = loggerConf.String("interval")
	r.level = loggerConf.String("level", "All")
	r.domain = loggerConf.String("domain", domain)
	r.server = loggerConf.String("server")
	r.service = loggerConf.String("service")
	r.layout, _ = loggerConf.GetSectionString("layout")
	r.rpcInvoker = rpc.NewInvoker(r.domain, r.server, address)
	return r, nil
}
func (r *rpcWriter) getConfig(rgst registry.Registry, path string) ([]byte, error) {
LOOP:
	for {
		select {
		case <-r.closeChan:
			break LOOP
		case <-time.After(time.Second):
			if b, err := rgst.Exists(path); err == nil && b {
				buff, _, err := rgst.GetValue(path)
				if err != nil {
					err = fmt.Errorf("无法获取RPC日志配置:%v", err)
					return nil, err
				}
				return buff, nil
			}
		}
	}
	return nil, fmt.Errorf("关闭监听:%s", path)

}
func (r *rpcWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = byte('[')
	p = append(p, byte(']'))
	dst := snappy.Encode(nil, p)
	str := fmt.Sprintf("%s", string(dst))
	_, _, _, err = r.rpcInvoker.Request(r.service, "GET", map[string]string{}, map[string]string{
		"__body": str,
	}, true)
	if err != nil && !r.writeError {
		r.writeError = true
		r.logger.Errorf("写入RPC日志失败:%v", err)
		return len(p) - 1, nil
	}
	if err == nil && r.writeError {
		r.writeError = false
		r.logger.Info("写入RPC日志成功")
	}
	return len(p) - 1, nil
}
func (r *rpcWriter) Close() error {
	r.once.Do(func() {
		close(r.closeChan)
		if r.rpcInvoker != nil {
			r.rpcInvoker.Close()
		}
	})
	return nil
}
