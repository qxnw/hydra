package web

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
	conterName := metrics.MakeName(ctx.tan.serverName+".request", metrics.COUNTER, "server", ctx.tan.ip, "client", client, "url", url)
	timerName := metrics.MakeName(ctx.tan.serverName+".request", metrics.TIMER, "server", ctx.tan.ip, "client", client, "url", url)
	successName := metrics.MakeName(ctx.tan.serverName+".success", metrics.METER, "server", ctx.tan.ip, "client", client, "url", url)
	failedName := metrics.MakeName(ctx.tan.serverName+".failed", metrics.METER, "server", ctx.tan.ip, "client", client, "url", url)

	counter := metrics.GetOrRegisterCounter(conterName, metrics.DefaultRegistry)
	counter.Inc(1)

	metrics.GetOrRegisterTimer(timerName, metrics.DefaultRegistry).Time(func() { m.execute(ctx) })
	counter.Dec(1)

	if !ctx.Written() {
		if ctx.Result == nil {
			ctx.Result = NotFound()
		}

		ctx.HandleError()
	}

	statusCode := ctx.Status()
	if statusCode >= 200 && statusCode < 400 {
		metrics.GetOrRegisterMeter(successName, metrics.DefaultRegistry).Mark(1)
	} else {
		metrics.GetOrRegisterMeter(failedName, metrics.DefaultRegistry).Mark(1)
	}
}
