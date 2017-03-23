package rpc

import "github.com/qxnw/lib4go/metrics"
import "github.com/qxnw/lib4go/concurrent/cmap"

//Limiter 流量限制组件
type Limiter struct {
	data cmap.ConcurrentMap
}

//NewLimiter 创建流程限制组件
func NewLimiter(lt map[string]int) *Limiter {
	m := &Limiter{}
	m.data = cmap.New()
	for k, v := range lt {
		m.data.Set(k, float64(v))
	}
	return m
}

//Update 更新限流规则
func (m *Limiter) Update(lt map[string]int) {
	m.data.Clear()
	for k, v := range lt {
		m.data.Set(k, float64(v))
	}
}

//Handle 限流处理
func (m *Limiter) Handle(ctx *Context) {
	service := ctx.Req().Service
	if count, ok := m.data.Get("*"); ok {
		limiterName := metrics.MakeName(ctx.server.serverName+".limiter", metrics.METER, "service", "*")
		meter := metrics.GetOrRegisterMeter(limiterName, metrics.DefaultRegistry)
		if meter.Rate1() >= count.(float64) {
			ctx.ServiceUnavailable()
			ctx.server.logger.Errorf("service:%s 超过总限流规则QPS %.0f/s", service, count)
			return
		}
		meter.Mark(1)
	}
	if count, ok := m.data.Get(service); ok {
		limiterName := metrics.MakeName(ctx.server.serverName+".limiter", metrics.METER, "service", service)
		meter := metrics.GetOrRegisterMeter(limiterName, metrics.DefaultRegistry)
		if meter.Rate1() >= count.(float64) {
			ctx.ServiceUnavailable()
			ctx.server.logger.Errorf("service:%s 超过服务限流规则QPS %.0f/s", service, count)
			return
		}
		meter.Mark(1)
	}
	ctx.Next()
}
