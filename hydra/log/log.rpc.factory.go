package log

import (
	"fmt"
	"path/filepath"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/logger"
)

//RemoteAppenderFactory 远程日志factory
type RemoteAppenderFactory struct {
	log      *logger.Logger
	writer   *RemoteLogWriter
	appender *logger.Appender
}

//NewRemoteAppenderFactory 创建远程日志factory
func NewRemoteAppenderFactory(rpcInvoker *rpc.Invoker, logger *logger.Logger) (f *RemoteAppenderFactory, err error) {
	f = &RemoteAppenderFactory{}
	f.writer, err = NewRemoteLogWriter(rpcInvoker, logger)
	return
}

//GetType 获取日志类型
func (f *RemoteAppenderFactory) GetType() string {
	return "rpc"
}

//MakeUniq 获取日志标识
func (f *RemoteAppenderFactory) MakeUniq(l *logger.Appender, event *logger.LogEvent) string {
	return "rpc"
}

//MakeAppender 构建日志组件
func (f *RemoteAppenderFactory) MakeAppender(l *logger.Appender, event *logger.LogEvent) (logger.IAppender, error) {
	rpc, err := NewRemoteAppender(f.writer, f.appender)
	return rpc, err
}

//ConfigRemoteLogger 配置远程日志组件
func ConfigRemoteLogger(platName string, systemName string, registryAddr string, log *logger.Logger) error {
	registry, err := registry.NewRegistryWithAddress(registryAddr, log)
	if err != nil {
		return err
	}
	conf := conf.NewRegistryConf(registry)
	path := filepath.Join("/", platName, "var", "logger", "logger")
	var nconf LogConf
	_, err = conf.GetObject(path, &nconf)
	if err != nil {
		err = fmt.Errorf("无法启用远程日志(%s):%v", path, err)
		return err
	}
	if b, err := govalidator.ValidateStruct(&nconf); !b {
		return err
	}
	layout, err := nconf.GetLayout()
	if err != nil {
		return err
	}
	rpcInvoker := rpc.NewInvoker(platName, systemName, registryAddr)
	f, err := NewRemoteAppenderFactory(rpcInvoker, log)
	if err != nil {
		return err
	}
	f.appender = &logger.Appender{Type: f.GetType(), Level: nconf.Level, Layout: layout, Interval: nconf.WriteInterval}
	if err != nil {
		return err
	}
	logger.RegistryFactory(f, f.appender)
	return nil
}
