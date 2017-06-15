package mq

import (
	"fmt"
	"sync"
	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
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
	logger          *logger.Logger
	reporter        *reporter
	registry        cmap.ConcurrentMap
	mu              sync.Mutex
	done            bool
	currentRegistry metrics.Registry
}

func NewInfluxMetric() *InfluxMetric {
	return &InfluxMetric{
		currentRegistry: metrics.NewRegistry(),
	}
}
func (m *InfluxMetric) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
}
func (m *InfluxMetric) RestartReport(host string, dataBase string, userName string, password string, timeSpan time.Duration, lg *logger.Logger) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
	m.logger = lg
	if m.logger == nil {
		m.logger = logger.GetSession("mq.metric", logger.CreateSession())
	}
	m.reporter = &reporter{Host: host, Database: dataBase, username: userName, password: password, timeSpan: timeSpan}
	m.reporter.influxdb, err = metrics.InfluxDB(m.currentRegistry, timeSpan, m.reporter.Host, m.reporter.Database, m.reporter.username, m.reporter.password, m.logger)
	if err != nil {
		return
	}
	go m.reporter.influxdb.Run()
	return nil
}

func (m *InfluxMetric) execute(task *Context) {
	task.Next()
}

//Handle 业务处理
func (m *InfluxMetric) Handle(ctx *Context) {
	processName := metrics.MakeName("mq.consumer.process", metrics.WORKING, "domain", ctx.server.domain, "name", ctx.server.serverName, "server", ctx.server.ip, "queue", ctx.queue)
	timerName := metrics.MakeName("mq.consumer.process", metrics.TIMER, "domain", ctx.server.domain, "name", ctx.server.serverName, "server", ctx.server.ip, "queue", ctx.queue)
	requestName := metrics.MakeName("mq.consumer.process", metrics.QPS, "domain", ctx.server.domain, "name", ctx.server.serverName, "server", ctx.server.ip, "queue", ctx.queue)
	metrics.GetOrRegisterRps(requestName, m.currentRegistry).Mark(1)

	process := metrics.GetOrRegisterCounter(processName, m.currentRegistry)
	process.Inc(1)
	metrics.GetOrRegisterTimer(timerName, m.currentRegistry).Time(func() { m.execute(ctx) })
	process.Dec(1)

	responseName := metrics.MakeName("mq.consumer.response", metrics.METER, "domain", ctx.server.domain, "name", ctx.server.serverName, "server",
		ctx.server.ip, "queue", ctx.queue, "status", fmt.Sprintf("%d", ctx.statusCode))
	metrics.GetOrRegisterMeter(responseName, m.currentRegistry).Mark(1)
}
