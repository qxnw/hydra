package monitor

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/metrics"
)

var registryMap cmap.ConcurrentMap
var influxdbCache cmap.ConcurrentMap
var log *logger.Logger

func init() {
	influxdbCache = cmap.New(2)
	registryMap = cmap.New(2)
	log = logger.New("monitor")
}
func getRegistry(cron string) metrics.Registry {
	_, r, _ := influxdbCache.SetIfAbsentCb(cron, func(i ...interface{}) (interface{}, error) {
		return metrics.NewRegistry(), nil
	}, cron)
	return r.(metrics.Registry)
}
func getReporter(c component.IContainer, ctx *context.Context, influxName string, lg *logger.Logger) (report metrics.IReporter, err error) {
	content, err := c.GetVarParam("influxdb", ctx.Request.Setting.GetString(influxName))
	if err != nil {
		return nil, err
	}
	cron := "@every 1m"
	if c, ok := ctx.Request.Ext.Get("__cron_"); ok {
		cron = c.(string)
	}
	key := fmt.Sprintf("%s-%s", cron, content)
	_, client, err := influxdbCache.SetIfAbsentCb(key, func(i ...interface{}) (interface{}, error) {

		content := i[0].(string)
		scron := i[1].(string)
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		userName := cnf.String("userName")
		password := cnf.String("password")
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		report, err := metrics.InfluxDB(getRegistry(scron), scron, host, dataBase, userName, password, lg)
		if err != nil {
			err = fmt.Errorf("创建inflxudb失败,err:%v", err)
			return nil, err
		}
		go report.Run()
		return report, err
	}, content, cron)
	if err != nil {
		return nil, err
	}
	return client.(metrics.IReporter), nil
}

func updateStatus(c component.IContainer, ctx *context.Context, influxName string, tagName string, value float64, params ...string) error {
	_, err := getReporter(c, ctx, influxName, log)
	if err != nil {
		return err
	}
	gaugeName := metrics.MakeName(tagName, metrics.GAUGE, params...) //堵塞计数
	cron := "@every 1m"
	if c, ok := ctx.Request.Ext.Get("__cron_"); ok {
		cron = c.(string)
	}
	metrics.GetOrRegisterGaugeFloat64(gaugeName, getRegistry(cron)).Update(value)
	return nil
}
func updateStatusInt64(c component.IContainer, ctx *context.Context, influxName string, tagName string, value int64, params ...string) error {
	_, err := getReporter(c, ctx, influxName, log)
	if err != nil {
		return err
	}
	gaugeName := metrics.MakeName(tagName, metrics.GAUGE, params...) //堵塞计数
	cron := "@every 1m"
	if c, ok := ctx.Request.Ext.Get("__cron_"); ok {
		cron = c.(string)
	}
	metrics.GetOrRegisterGauge(gaugeName, getRegistry(cron)).Update(value)
	return nil
}

func updateCPUStatus(c component.IContainer, ctx *context.Context, value float64, params ...string) error {
	return updateStatus(c, ctx, "influxdb", "monitor.cpu.status", value, params...)
}
func updateMemStatus(c component.IContainer, ctx *context.Context, value float64, params ...string) error {
	return updateStatus(c, ctx, "influxdb", "monitor.mem.status", value, params...)
}
func updateDiskStatus(c component.IContainer, ctx *context.Context, value float64, params ...string) error {
	return updateStatus(c, ctx, "influxdb", "monitor.disk.status", value, params...)
}
func updateHTTPStatus(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.http.status", value, params...)
}
func updateTCPStatus(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.tcp.status", value, params...)
}
func updateRegistryStatus(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.registry.status", value, params...)
}
func updateDBStatus(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.db.status", value, params...)
}

func updateNetRecvStatus(c component.IContainer, ctx *context.Context, value uint64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.net.recv", int64(value), params...)
}
func updateNetSentStatus(c component.IContainer, ctx *context.Context, value uint64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.net.sent", int64(value), params...)
}
func updateNetConnectCountStatus(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.net.conn", value, params...)
}
func updateNginxErrorCount(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.nginx.error", value, params...)
}
func updateNginxAccessCount(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.nginx.access", value, params...)
}
func updateredisListCount(c component.IContainer, ctx *context.Context, value int64, params ...string) error {
	return updateStatusInt64(c, ctx, "influxdb", "monitor.queue.count", value, params...)
}
