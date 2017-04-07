package context

//EngineHandler context handle
type EngineHandler interface {
	Handle(name string, method string, service string, params string, c Context) (*Response, error)
}

//Context 服务输出及Task执行的上下文
type Context map[string]interface{}

//Response 响应
type Response struct {
	Content string
	Status  int
	Params  map[string]interface{}
}
