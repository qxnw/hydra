package rpc

import (
	"time"

	"github.com/qxnw/lib4go/metrics"
)

//InfluxMetric metric组件
type InfluxMetric struct {
	Host     string
	Database string
	username string
	password string
	timeSpan time.Duration
}

//NewInfluxMetric 创建基于influxdb的metric组件
func NewInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) *InfluxMetric {
	m := &InfluxMetric{Host: host, Database: dataBase, username: userName, password: password, timeSpan: timeSpan}
	go metrics.InfluxDB(metrics.DefaultRegistry, timeSpan, m.Host, m.Database,
		m.username,
		m.password)
	go metrics.DefaultRegistry.RunHealthchecks()
	return m
}

func (m *InfluxMetric) execute(context *Context) {
	if action := context.Action(); action != nil {
		if l, ok := action.(LogInterface); ok {
			l.SetLogger(context.Logger)
		}
	}
	context.Next()
}

//Handle 业务处理
func (m *InfluxMetric) Handle(ctx *Context) {
	service := ctx.Req().Service
	client := ctx.IP()
	processName := metrics.MakeName(ctx.server.serverName+".process", metrics.COUNTER, "server", ctx.server.address, "client", client, "service", service)
	timerName := metrics.MakeName(ctx.server.serverName+".request", metrics.TIMER, "server", ctx.server.address, "client", client, "service", service)

	totalName := metrics.MakeName(ctx.server.serverName+".request", metrics.METER, "server", ctx.server.address, "client", client, "service", service)
	successName := metrics.MakeName(ctx.server.serverName+".success", metrics.METER, "server", ctx.server.address, "client", client, "service", service)
	failedName := metrics.MakeName(ctx.server.serverName+".failed", metrics.METER, "server", ctx.server.address, "client", client, "service", service)

	process := metrics.GetOrRegisterCounter(processName, metrics.DefaultRegistry)
	process.Inc(1)
	metrics.GetOrRegisterMeter(totalName, metrics.DefaultRegistry).Mark(1)
	metrics.GetOrRegisterTimer(timerName, metrics.DefaultRegistry).Time(func() { m.execute(ctx) })
	process.Dec(1)

	if !ctx.Written() {
		if ctx.Result == nil {
			ctx.Result = NotFound()
		}
		ctx.HandleError()
	}
	statusCode := ctx.Writer.Code

	statusName := metrics.MakeName(ctx.server.serverName+".status", metrics.METER, "server", ctx.server.address, "client", client, "service", service, "status", string(statusCode))
	metrics.GetOrRegisterMeter(statusName, metrics.DefaultRegistry).Mark(1)

	if statusCode >= 200 && statusCode < 400 {
		metrics.GetOrRegisterMeter(successName, metrics.DefaultRegistry).Mark(1)
	} else {
		metrics.GetOrRegisterMeter(failedName, metrics.DefaultRegistry).Mark(1)
	}
}
