package circuit

import (
	"sync/atomic"
	"time"
)

type option struct {
	RPS         int
	EPPS        int
	RJTPS       int
	Timeout     int
	SleepWindow int64
}

//Option 配置选项
type Option func(*option)

//WithRPS 每秒请求数
func WithRPS(i int) Option {
	return func(o *option) {
		o.RPS = i
	}
}

//WithEPPS 每秒失败数
func WithEPPS(i int) Option {
	return func(o *option) {
		o.EPPS = i
	}
}

//WithReject 每秒拒绝访问数
func WithReject(i int) Option {
	return func(o *option) {
		o.RJTPS = i
	}
}

//WithTimeout 每秒超时请求数
func WithTimeout(i int) Option {
	return func(o *option) {
		o.Timeout = i
	}
}

//WithSleepWindow 每秒失败数
func WithSleepWindow(i int64) Option {
	return func(o *option) {
		o.SleepWindow = i * int64(time.Millisecond)
	}
}

//CircuitBreaker 熔断管理
type CircuitBreaker struct {
	Name                   string
	open                   int32
	forceOpen              int32
	openedOrLastTestedTime int64
	metrics                *StandardMetricCollector
	*option
}

// NewCircuitBreaker creates a CircuitBreaker with associated Health
func NewCircuitBreaker(opts ...Option) *CircuitBreaker {
	c := &CircuitBreaker{
		metrics: NewStandardMetricCollector(),
		option: &option{
			RPS:         0,
			EPPS:        -1,
			RJTPS:       -1,
			SleepWindow: 0,
		},
	}
	for _, opt := range opts {
		opt(c.option)
	}
	return c
}

// ToggleForceOpen allows manually causing the fallback logic for all instances
// of a given command.
func (circuit *CircuitBreaker) ToggleForceOpen(toggle bool) {
	if toggle {
		circuit.forceOpen = 0
		return
	}
	circuit.forceOpen = -1
}

// IsOpen is called before any Command execution to check whether or
// not it should be attempted. An "open" circuit means it is disabled.
func (circuit *CircuitBreaker) IsOpen() bool {
	o := circuit.forceOpen == 0 || circuit.open == 0
	if o {
		return true
	}
	now := time.Now()
	if circuit.RPS == 0 || circuit.metrics.NumRequests().Sum(now) < uint64(circuit.RPS) {
		return false
	}

	if !circuit.IsHealthy(now) {
		circuit.setOpen()
		return true
	}
	return false
}

// AllowRequest is checked before a command executes, ensuring that circuit state and metric health allow it.
// When the circuit is open, this call will occasionally return true to measure whether the external service
// has recovered.
func (circuit *CircuitBreaker) AllowRequest() bool {
	return !circuit.IsOpen() || circuit.allowSingleTest()
}

func (circuit *CircuitBreaker) allowSingleTest() bool {
	if circuit.SleepWindow == 0 {
		return true
	}
	now := time.Now().UnixNano()
	openedOrLastTestedTime := atomic.LoadInt64(&circuit.openedOrLastTestedTime)
	if circuit.open == 0 && now > openedOrLastTestedTime+circuit.SleepWindow {
		swapped := atomic.CompareAndSwapInt64(&circuit.openedOrLastTestedTime, openedOrLastTestedTime, now)
		return swapped
	}
	return false
}

func (circuit *CircuitBreaker) setOpen() {
	if atomic.CompareAndSwapInt32(&circuit.open, -1, 0) {
		circuit.openedOrLastTestedTime = time.Now().UnixNano()
	}
}

func (circuit *CircuitBreaker) setClose() {
	if atomic.CompareAndSwapInt32(&circuit.open, 0, -1) {
		circuit.metrics.Reset()
	}
}

//IsHealthy 当前服务器健康状况
func (circuit *CircuitBreaker) IsHealthy(t time.Time) bool {
	return (circuit.EPPS < 0 || circuit.metrics.FailurePercent(t) > circuit.EPPS) && (circuit.RJTPS < 0 || circuit.metrics.RejectPercent(t) > circuit.RJTPS)
}

// ReportEvent records command metrics for tracking recent error rates and exposing data to the dashboard.
func (circuit *CircuitBreaker) ReportEvent(event string, i uint64) error {
	if event == EventSuccess && circuit.open == 0 {
		circuit.setClose()
	}
	switch event {
	case EventFailure:
		circuit.metrics.Failure(i)
	case EventFallbackFailure:
		circuit.metrics.FallbackFailure(i)
	case EventFallbackSuccess:
		circuit.metrics.FallbackSuccess(i)
	case EventReject:
		circuit.metrics.Reject(i)
	case EventShortCircuit:
		circuit.metrics.ShortCircuit(i)
	case EventSuccess:
		circuit.metrics.Success(i)

	}
	return nil
}

var (
	EventSuccess      = "SUCCESS"
	EventFailure      = "FAILURE"
	EventReject       = "REJECT"
	EventShortCircuit = "SHORT_CIRCUIT"

	EventFallbackSuccess = "FALLBACK_SUCCESS"
	EventFallbackFailure = "FALLBACK_FAILURE"
)
