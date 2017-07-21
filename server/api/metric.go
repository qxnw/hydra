package api

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
}
type InfluxMetric struct {
	logger          *logger.Logger
	reporter        *reporter
	registry        cmap.ConcurrentMap
	mu              sync.Mutex
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

func (m *InfluxMetric) RestartReport(host string, dataBase string, userName string, password string, timeSpan time.Duration,
	lg *logger.Logger) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
	m.logger = lg
	if m.logger == nil {
		m.logger = logger.GetSession("api.metric", logger.CreateSession())
	}
	m.reporter = &reporter{Host: host, Database: dataBase, username: userName, password: password, timeSpan: timeSpan}
	m.reporter.influxdb, err = metrics.InfluxDB(m.currentRegistry,
		timeSpan,
		m.reporter.Host, m.reporter.Database,
		m.reporter.username,
		m.reporter.password, m.logger)
	if err != nil {
		return
	}
	go m.reporter.influxdb.Run()
	return nil
}

func (m *InfluxMetric) execute(context *Context) {
	if action := context.Action(); action != nil {
		if l, ok := action.(LogInterface); ok {
			l.SetLogger(context.Logger)
		}
	}
	context.Next()
}

func (m *InfluxMetric) Handle(ctx *Context) {
	url := ctx.Req().URL.Path
	client := ctx.IP()
	conterName := metrics.MakeName("api.server.request", metrics.WORKING, "domain", ctx.Server.domain, "name", ctx.Server.serverName, "server", ctx.Server.ip, "client", client, "url", url) //堵塞计数
	timerName := metrics.MakeName("api.server.request", metrics.TIMER, "domain", ctx.Server.domain, "name", ctx.Server.serverName, "server", ctx.Server.ip, "client", client, "url", url)    //堵塞计数
	requestName := metrics.MakeName("api.server.request", metrics.QPS, "domain", ctx.Server.domain, "name", ctx.Server.serverName, "server", ctx.Server.ip,
		"client", client, "url", url) //请求数
	metrics.GetOrRegisterRps(requestName, m.currentRegistry).Mark(1)

	counter := metrics.GetOrRegisterCounter(conterName, m.currentRegistry)
	counter.Inc(1)
	metrics.GetOrRegisterTimer(timerName, m.currentRegistry).Time(func() { m.execute(ctx) })
	counter.Dec(1)

	if !ctx.Written() {
		if ctx.Result == nil {
			ctx.Result = NotFound()
		}
		ctx.HandleError()
	}

	statusCode := ctx.Status()
	responseName := metrics.MakeName("api.server.response", metrics.METER, "domain", ctx.Server.domain, "name", ctx.Server.serverName, "server", ctx.Server.ip,
		"client", client, "url", url, "status", fmt.Sprintf("%d", statusCode)) //完成数
	metrics.GetOrRegisterMeter(responseName, m.currentRegistry).Mark(1)

}
