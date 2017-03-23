package balancer

import "github.com/qxnw/lib4go/metrics"
import "github.com/qxnw/lib4go/concurrent/cmap"

//Limiter 流量限制组件
type Limiter struct {
	data     cmap.ConcurrentMap
	service  string
	registry metrics.Registry
}

//NewLimiter 创建流程限制组件
func NewLimiter(service string, lt map[string]int) *Limiter {
	m := &Limiter{service: service}
	m.registry = metrics.NewRegistry()
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

//Check 检查当前服务是否限流
func (m *Limiter) Check(ip string) bool {
	if count, ok := m.data.Get("*"); ok {
		limiterName := metrics.MakeName(".limiter", metrics.METER, "service", m.service)
		meter := metrics.GetOrRegisterMeter(limiterName, m.registry)
		if meter.Rate1() >= count.(float64) {
			return false
		}
		meter.Mark(1)
	}
	if count, ok := m.data.Get(m.service); ok {
		limiterName := metrics.MakeName(".limiter", metrics.METER, "service", m.service, "ip", ip)
		meter := metrics.GetOrRegisterMeter(limiterName, m.registry)
		if meter.Rate1() >= count.(float64) {
			return false
		}
		meter.Mark(1)
	}
	return true
}
