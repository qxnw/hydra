package context

type circuitBreakerParam struct {
	*inputParams
	ext map[string]interface{}
}

//IsOpen 熔断开发是否打开
func (s *circuitBreakerParam) IsOpen() bool {
	return true
}
