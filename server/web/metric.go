package web

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
	conterName := metrics.MakeName(ctx.tan.serverName+".request", metrics.WORKING, "server", ctx.tan.ip, "client", client, "url", url) //堵塞计数
	timerName := metrics.MakeName(ctx.tan.serverName+".response", metrics.TIMER, "server", ctx.tan.ip, "client", client, "url", url)   //响应时长

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
	responseName := metrics.MakeName(ctx.tan.serverName+".response", metrics.METER, "server", ctx.tan.ip,
		"client", client, "url", url, "status", fmt.Sprintf("%d", statusCode)) //响应状态码
	metrics.GetOrRegisterMeter(responseName, metrics.DefaultRegistry).Mark(1)
}
