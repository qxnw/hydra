package mq

import (
	"sync"
	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/metrics"
)

type reporter struct {
	influxdb metrics.IReporter
	Host     string
	Database string
	username string
	password string
	timeSpan time.Duration
	done     bool
}
type InfluxMetric struct {
	reporter *reporter
	registry cmap.ConcurrentMap
	mu       sync.Mutex
}

func NewInfluxMetric() *InfluxMetric {
	return &InfluxMetric{}
}
func (m *InfluxMetric) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
}
func (m *InfluxMetric) RestartReport(host string, dataBase string, userName string, password string, timeSpan time.Duration) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
	m.reporter = &reporter{Host: host, Database: dataBase, username: userName, password: password, timeSpan: timeSpan}
	m.reporter.influxdb, err = metrics.InfluxDB(metrics.DefaultRegistry, timeSpan, m.reporter.Host, m.reporter.Database, m.reporter.username, m.reporter.password)
	if err != nil {
		return
	}
	go m.reporter.influxdb.Run()
	go metrics.DefaultRegistry.RunHealthchecks()
	return nil
}

func (m *InfluxMetric) execute(task *Context) {
	task.Next()
}

//Handle 业务处理
func (m *InfluxMetric) Handle(ctx *Context) {
	service := ctx.taskName
	processName := metrics.MakeName(ctx.server.serverName+".process", metrics.COUNTER, "server", ctx.server.ip, "service", service)
	timerName := metrics.MakeName(ctx.server.serverName+".request", metrics.TIMER, "server", ctx.server.ip, "service", service)

	totalName := metrics.MakeName(ctx.server.serverName+".request", metrics.METER, "server", ctx.server.ip, "service", service)
	successName := metrics.MakeName(ctx.server.serverName+".success", metrics.METER, "server", ctx.server.ip, "service", service)
	failedName := metrics.MakeName(ctx.server.serverName+".failed", metrics.METER, "server", ctx.server.ip, "service", service)

	process := metrics.GetOrRegisterCounter(processName, metrics.DefaultRegistry)
	process.Inc(1)
	metrics.GetOrRegisterMeter(totalName, metrics.DefaultRegistry).Mark(1)
	metrics.GetOrRegisterTimer(timerName, metrics.DefaultRegistry).Time(func() { m.execute(ctx) })
	process.Dec(1)

	if ctx.err == nil {
		metrics.GetOrRegisterMeter(successName, metrics.DefaultRegistry).Mark(1)
	} else {
		metrics.GetOrRegisterMeter(failedName, metrics.DefaultRegistry).Mark(1)
	}
}
