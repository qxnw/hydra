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

func updateStatus(ctx *context.Context, influxName string, serverName string, tagName string, value float64) error {
	_, err := getReporter(ctx, influxName, log)
	if err != nil {
		return err
	}
	gaugeName := metrics.MakeName(tagName, metrics.GAUGEFLOAST64, "server", serverName) //堵塞计数
	metrics.GetOrRegisterGaugeFloat64(gaugeName, currentRegistry).Update(value)
	return nil
}

func updateCPUStatus(ctx *context.Context, influxName string, serverName string, value float64) error {
	return updateStatus(ctx, influxName, serverName, "monitor.cpu.status", value)
}
func updateMemStatus(ctx *context.Context, influxName string, serverName string, value float64) error {
	return updateStatus(ctx, influxName, serverName, "monitor.mem.status", value)
}
