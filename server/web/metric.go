package web

import (
	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/metrics"
)

type InfluxMetric struct {
	Host     string
	Database string
	username string
	password string
	timeSpan time.Duration
	registry cmap.ConcurrentMap
}

func NewInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) *InfluxMetric {
	m := &InfluxMetric{Host: host, Database: dataBase, username: userName, password: password, timeSpan: timeSpan}
	go metrics.InfluxDB(metrics.DefaultRegistry, timeSpan, m.Host, m.Database,
		m.username,
		m.password)
	go metrics.DefaultRegistry.RunHealthchecks()
	return m
}

func (m *InfluxMetric) Execute(context *Context) {
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

	metrics.GetOrRegisterTimer(timerName, metrics.DefaultRegistry).Time(func() { m.Execute(ctx) })
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
