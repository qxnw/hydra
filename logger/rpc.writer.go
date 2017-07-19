package logger

import (
	"fmt"

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
	rpcInvoker *rpc.RPCInvoker
	*rpcLoggerSetting
}

func newRPCWriter(domain string, address string, logger *logger.Logger) (r *rpcWriter, err error) {
	r = &rpcWriter{
		cdomain:          domain,
		address:          address,
		rpcLoggerSetting: &rpcLoggerSetting{},
	}
	registry, err := registry.NewRegistryWithAddress(address, logger)
	if err != nil {
		err = fmt.Errorf("初始化注册中心失败：%s:%v", address, err)
		return nil, err
	}
	path := fmt.Sprintf("%s/var/global/logger", domain)
	if b, err := registry.Exists(path); !b || err != nil {
		return nil, fmt.Errorf("rpc日志未配置:%v", err)
	}
	buff, _, err := registry.GetValue(path)
	if err != nil {
		err = fmt.Errorf("无法获取RPC日志配置:%v", err)
		return nil, err
	}
	loggerConf, err := conf.NewJSONConfWithJson(string(buff), 0, nil, nil)
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
	r.rpcInvoker = rpc.NewRPCInvoker(r.domain, r.server, address)
	return r, nil
}

func (r *rpcWriter) Write(p []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			n = len(p) - 1
		}
	}()
	p[0] = byte('[')
	p = append(p, byte(']'))
	dst := snappy.Encode(nil, p)
	str := fmt.Sprintf("%s", string(dst))
	input := map[string]string{
		"__body": str,
	}
	_, _, _, err = r.rpcInvoker.Request(r.service, input, true)
	if err != nil {
		fmt.Println(err)
		return len(p), nil
	}
	return len(p) - 1, nil
}
func (r *rpcWriter) Close() error {
	if r.rpcInvoker != nil {
		r.rpcInvoker.Close()
	}

	return nil
}
