package middleware

import (
	"fmt"
	"sync"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
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
	cron     string
}

//Metric 服务器处理能力统计
type Metric struct {
	logger          *logger.Logger
	reporter        *reporter
	registry        cmap.ConcurrentMap
	mu              sync.Mutex
	currentRegistry metrics.Registry
	conf            *conf.ServerConf
}

//NewMetric new metric
func NewMetric(conf *conf.ServerConf) *Metric {
	return &Metric{
		conf:            conf,
		currentRegistry: metrics.NewRegistry(),
	}
}

//Stop stop metric
func (m *Metric) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
}

//Restart restart metric
func (m *Metric) Restart(host string, dataBase string, userName string, password string, cron string,
	lg *logger.Logger) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reporter != nil && m.reporter.influxdb != nil {
		m.reporter.influxdb.Close()
	}
	m.logger = lg
	m.reporter = &reporter{Host: host, Database: dataBase, username: userName, password: password, cron: cron}
	m.reporter.influxdb, err = metrics.InfluxDB(m.currentRegistry,
		cron,
		m.reporter.Host, m.reporter.Database,
		m.reporter.username,
		m.reporter.password, m.logger)
	if err != nil {
		return
	}
	go m.reporter.influxdb.Run()
	return nil
}

//Handle 处理请求
func (m *Metric) Handle() dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		url := ctx.Request.GetService()

		conterName := metrics.MakeName(m.conf.Type+".server.request", metrics.WORKING, "domain", m.conf.Domain, "server", m.conf.Name, "cluster", m.conf.Cluster, "ip", m.conf.IP, "url", url) //堵塞计数
		timerName := metrics.MakeName(m.conf.Type+".server.request", metrics.TIMER, "domain", m.conf.Domain, "server", m.conf.Name, "cluster", m.conf.Cluster, "ip", m.conf.IP, "url", url)    //堵塞计数
		requestName := metrics.MakeName(m.conf.Type+".server.request", metrics.QPS, "domain", m.conf.Domain, "server", m.conf.Name, "cluster", m.conf.Cluster, "ip", m.conf.IP, "url", url)    //请求数

		metrics.GetOrRegisterQPS(requestName, m.currentRegistry).Mark(1)

		counter := metrics.GetOrRegisterCounter(conterName, m.currentRegistry)
		counter.Inc(1)
		metrics.GetOrRegisterTimer(timerName, m.currentRegistry).Time(func() { ctx.Next() })
		counter.Dec(1)

		statusCode := ctx.Writer.Status()
		responseName := metrics.MakeName(m.conf.Type+".server.response", metrics.METER, "domain", m.conf.Domain, "server", m.conf.Name, "cluster", m.conf.Cluster, "ip", m.conf.IP,
			"url", url, "status", fmt.Sprintf("%d", statusCode)) //完成数
		metrics.GetOrRegisterMeter(responseName, m.currentRegistry).Mark(1)
	}

}
