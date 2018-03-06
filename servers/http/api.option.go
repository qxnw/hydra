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
	metric *middleware.Metric
	static *middleware.StaticOptions
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

//WithStatic 设置静态文件目录
func WithStatic(enable bool, prefix string, dir string, listDir bool, exts []string) Option {
	return func(o *option) {
		if o.static == nil {
			o.static = &middleware.StaticOptions{
				Enable:   enable,
				Prefix:   prefix,
				RootPath: dir,
				Exts:     exts,
			}
			return
		}
		o.static.Enable = enable
		o.static.Prefix = prefix
		o.static.RootPath = dir
		o.static.Exts = exts
	}
}
