package standard

import (
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
	"github.com/qxnw/hydra/servers/pkg/middleware"
	"github.com/qxnw/lib4go/logger"
)

type Handler interface {
	Handle(*dispatcher.Context)
}
type option struct {
	ip string
	*logger.Logger
	metric *middleware.Metric
}

//Option 配置选项
type Option func(*option)

//WithLogger 设置日志记录组件
func WithLogger(logger *logger.Logger) Option {
	return func(o *option) {
		o.Logger = logger
	}
}

//WithIP 设置ip地址
func WithIP(ip string) Option {
	return func(o *option) {
		o.ip = ip
	}
}

//WithMetric 设置基于influxdb的系统监控组件
func WithMetric(host string, dataBase string, userName string, password string, cron string) Option {
	return func(o *option) {
		o.metric.Restart(host, dataBase, userName, password, cron, o.Logger)
	}
}
