package cron

import (
	"fmt"
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
	return nil
}

func (m *InfluxMetric) execute(task *Task) {
	task.DoNext()
}

//Handle 业务处理
func (m *InfluxMetric) Handle(ctx *Task) {

	conterName := metrics.MakeName(ctx.server.serverName+".process", metrics.WORKING, "server", ctx.server.ip, "name", ctx.taskName)
	timerName := metrics.MakeName(ctx.server.serverName+".request", metrics.TIMER, "server", ctx.server.ip, "name", ctx.taskName)

	process := metrics.GetOrRegisterCounter(conterName, metrics.DefaultRegistry)
	process.Inc(1)
	metrics.GetOrRegisterTimer(timerName, metrics.DefaultRegistry).Time(func() { m.execute(ctx) })
	process.Dec(1)

	responseName := metrics.MakeName(ctx.server.serverName+".response", metrics.METER, "server",
		ctx.server.ip, "name", ctx.taskName, "status", fmt.Sprintf("%d", ctx.statusCode))
	metrics.GetOrRegisterMeter(responseName, metrics.DefaultRegistry).Mark(1)
}
