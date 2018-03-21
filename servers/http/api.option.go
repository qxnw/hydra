package http

import (
	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/lib4go/logger"
)

type Handler interface {
	Handle(*gin.Context)
}
type option struct {
	ip string
	*logger.Logger
	readTimeout       int
	writeTimeout      int
	readHeaderTimeout int
	metric            *middleware.Metric
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

//WithTimeout 设置服务器超时时长
func WithTimeout(readTimeout int, writeTimeout int, readHeaderTimeout int) Option {
	return func(o *option) {
		o.readTimeout = readTimeout
		o.writeTimeout = writeTimeout
		o.readHeaderTimeout = readHeaderTimeout
	}
}

//WithMetric 设置基于influxdb的系统监控组件
func WithMetric(host string, dataBase string, userName string, password string, cron string) Option {
	return func(o *option) {
		o.metric.Restart(host, dataBase, userName, password, cron, o.Logger)
	}
}
