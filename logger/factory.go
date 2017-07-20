package logger

import "github.com/qxnw/lib4go/logger"

type RPCAppenderFactory struct {
	log      *logger.Logger
	writer   *rpcWriter
	appender *logger.Appender
}

func NewRPCAppenderFactory(domain string, address string, log *logger.Logger) (f *RPCAppenderFactory, err error) {
	f = &RPCAppenderFactory{}
	f.writer, err = newRPCWriter(domain, address, log)
	return
}

func (f *RPCAppenderFactory) GetType() string {
	return "rpc"
}
func (f *RPCAppenderFactory) MakeUniq(l *logger.Appender, event *logger.LogEvent) string {
	return "rpc"
}

func (f *RPCAppenderFactory) MakeAppender(l *logger.Appender, event *logger.LogEvent) (logger.IAppender, error) {
	rpc, err := NewRPCAppender(f.writer, f.appender)
	return rpc, err
}
func ConfigRPCLogger(domain string, address string, log *logger.Logger) error {
	f, err := NewRPCAppenderFactory(domain, address, log)
	if err != nil {
		return err
	}
	f.appender = &logger.Appender{Type: "rpc", Level: f.writer.level, Layout: f.writer.layout, Interval: f.writer.interval}
	if err != nil {
		return err
	}
	logger.RegistryFactory(f, f.appender)
	return nil
}
