package api

type webRouter struct {
	Method      []string
	Path        string
	Handler     interface{}
	Middlewares []Handler
	params      map[string]string
}

//Get 设置get路由
func (t *WebServer) Get(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"GET", "HEAD:Get"}, url, c, middlewares...)
}

//Post 设置Post路由
func (t *WebServer) Post(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"POST"}, url, c, middlewares...)
}

//Head 设置Head路由
func (t *WebServer) Head(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"HEAD"}, url, c, middlewares...)
}

//Options 设置Options路由
func (t *WebServer) Options(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"OPTIONS"}, url, c, middlewares...)
}

//Trace 设置Trace路由
func (t *WebServer) Trace(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"TRACE"}, url, c, middlewares...)
}

//Patch 设置Patch路由
func (t *WebServer) Patch(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"PATCH"}, url, c, middlewares...)
}

//Delete 设置Delete路由
func (t *WebServer) Delete(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"DELETE"}, url, c, middlewares...)
}

//Put 设置Put路由
func (t *WebServer) Put(url string, c interface{}, middlewares ...Handler) {
	t.Route([]string{"PUT"}, url, c, middlewares...)
}

//Any 设置Any路由
func (t *WebServer) Any(url string, c interface{}, middlewares ...Handler) {
	t.Route(SupportMethods, url, c, middlewares...)
	t.Route([]string{"HEAD:Get"}, url, c, middlewares...)
}

//Use 使用新的插件
func (t *WebServer) Use(handlers ...Handler) {
	t.handlers = append(t.handlers, handlers...)
}
