package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/metrics"
)

var currentRegistry metrics.Registry
var influxdbCache cmap.ConcurrentMap
var log *logger.Logger

func init() {
	influxdbCache = cmap.New(2)
	currentRegistry = metrics.NewRegistry()
	log = logger.New("monitor")
}

func getReporter(ctx *context.Context, influxName string, lg *logger.Logger) (report metrics.IReporter, err error) {
	content, err := ctx.Input.GetVarParamByArgsName("influxdb", influxName)
	if err != nil {
		return nil, err
	}
	_, client, err := influxdbCache.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		userName := cnf.String("userName")
		password := cnf.String("password")
		timeSpan, _ := cnf.Int("span", 60)
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		report, err := metrics.InfluxDB(currentRegistry,
			time.Second*time.Duration(timeSpan),
			host, dataBase, userName, password, lg)
		if err != nil {
			err = fmt.Errorf("创建inflxudb失败,err:%v", err)
			return nil, err
		}
		go report.Run()
		return report, err
	})
	if err != nil {
		return nil, err
	}
	return client.(metrics.IReporter), nil
}

func updateStatus(ctx *context.Context, influxName string, tagName string, value float64, params ...string) error {
	_, err := getReporter(ctx, influxName, log)
	if err != nil {
		return err
	}
	gaugeName := metrics.MakeName(tagName, metrics.GAUGE, params...) //堵塞计数
	metrics.GetOrRegisterGaugeFloat64(gaugeName, currentRegistry).Update(value)
	return nil
}
func updateStatusInt64(ctx *context.Context, influxName string, tagName string, value int64, params ...string) error {
	_, err := getReporter(ctx, influxName, log)
	if err != nil {
		return err
	}
	gaugeName := metrics.MakeName(tagName, metrics.GAUGE, params...) //堵塞计数
	metrics.GetOrRegisterGauge(gaugeName, currentRegistry).Update(value)
	return nil
}

func updateCPUStatus(ctx *context.Context, influxName string, value float64, params ...string) error {
	return updateStatus(ctx, influxName, "monitor.cpu.status", value, params...)
}
func updateMemStatus(ctx *context.Context, influxName string, value float64, params ...string) error {
	return updateStatus(ctx, influxName, "monitor.mem.status", value, params...)
}
func updateDiskStatus(ctx *context.Context, influxName string, value float64, params ...string) error {
	return updateStatus(ctx, influxName, "monitor.disk.status", value, params...)
}
func updateHTTPStatus(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.http.status", value, params...)
}
func updateTCPStatus(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.tcp.status", value, params...)
}
func updateRegistryStatus(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.registry.status", value, params...)
}
func updateDBStatus(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.db.status", value, params...)
}

func updateNetRecvStatus(ctx *context.Context, influxName string, value uint64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.net.recv", int64(value), params...)
}
func updateNetSentStatus(ctx *context.Context, influxName string, value uint64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.net.sent", int64(value), params...)
}
func updateNetConnectCountStatus(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.net.conn", value, params...)
}
func updateNginxErrorCount(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.nginx.error", value, params...)
}
func updateNginxQPSCount(ctx *context.Context, influxName string, value int64, params ...string) error {
	return updateStatusInt64(ctx, influxName, "monitor.nginx.qps", value, params...)
}
