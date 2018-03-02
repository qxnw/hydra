package circuit

import (
	"sync"

	"github.com/qxnw/hydra/servers/pkg/conf"
)

type NamedCircuitBreakers struct {
	breakers      sync.Map
	conf          *conf.CircuitBreaker
	closedBreaker *CircuitBreaker
	openedBreaker *CircuitBreaker
}

func NewNamedCircuitBreakers(conf *conf.CircuitBreaker) *NamedCircuitBreakers {
	n := &NamedCircuitBreakers{
		conf:          conf,
		closedBreaker: NewCircuitBreaker(),
		openedBreaker: NewCircuitBreaker(WithSleepWindow(0)),
	}
	n.openedBreaker.ToggleForceOpen(true)
	return n
}
func (c *NamedCircuitBreakers) getBreakerConf(url string) *conf.Breaker {
	if breakerConf, ok := c.conf.CircuitBreakers[url]; ok {
		return breakerConf
	}
	if breakerConf, ok := c.conf.CircuitBreakers["*"]; ok {
		return breakerConf
	}
	return nil
}

//GetBreaker 获取当前URL的熔断信息
func (c *NamedCircuitBreakers) GetBreaker(url string) *CircuitBreaker {
	if !c.conf.Enable {
		return c.closedBreaker
	}
	if c.conf.ForceBreak {
		return c.openedBreaker
	}
	var conf *conf.Breaker
	if conf = c.getBreakerConf(url); conf == nil {
		return c.closedBreaker
	}
	breaker, _ := c.breakers.LoadOrStore(conf.URL, NewCircuitBreaker(
		WithFPPS(conf.FailedPercent),
		WithRPS(conf.RequestPerSecond),
		WithReject(conf.RejectPerSecond),
		WithSleepWindow(int64(c.conf.SwitchWindow)),
	))
	cb := breaker.(*CircuitBreaker)
	cb.ToggleForceOpen(c.conf.ForceBreak)
	return cb
}

//Close 关闭熔断配置
func (c *NamedCircuitBreakers) Close() {
	c.conf.Enable = false
	c.conf.ForceBreak = false
}
