package web

import (
	"net/http"
	"strconv"
	"strings"
)

type HandlerFunc func(ctx *Context)

func (h HandlerFunc) Handle(ctx *Context) {
	h(ctx)
}
func WrapBefore(handler http.Handler) HandlerFunc {
	return func(ctx *Context) {
		handler.ServeHTTP(ctx.ResponseWriter, ctx.Req())
		ctx.Next()
	}
}

func WrapAfter(handler http.Handler) HandlerFunc {
	return func(ctx *Context) {
		ctx.Next()
		handler.ServeHTTP(ctx.ResponseWriter, ctx.Req())
	}
}

func (t *WebServer) UseHandler(handler http.Handler) {
	t.Use(WrapBefore(handler))
}
func (t *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	resp := t.respPool.Get().(*responseWriter)
	resp.reset(w)

	ctx := t.ctxPool.Get().(*Context)
	ctx.tan = t
	ctx.reset(req, resp)

	ctx.invoke()
	// if there is no logging or error handle, so the last written check.
	if !ctx.Written() {
		p := req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			p = p + "?" + req.URL.RawQuery
		}

		if ctx.Route() != nil {
			if ctx.Result == nil {
				ctx.WriteString("")
				t.logger.Info(req.Method, ctx.Status(), p)
				ctx.Close()
				t.ctxPool.Put(ctx)
				t.respPool.Put(resp)
				return
			}
			panic("result should be handler before")
		}

		if ctx.Result == nil {
			ctx.Result = NotFound()
		}

		ctx.HandleError()

		t.logger.Error(req.Method, ctx.Status(), p)
	}
	ctx.Close()
	t.ctxPool.Put(ctx)
	t.respPool.Put(resp)
}

func (t *WebServer) getAddress(args ...interface{}) string {
	var host string
	var port int

	if len(args) == 1 {
		switch arg := args[0].(type) {
		case string:
			addrs := strings.Split(args[0].(string), ":")
			if len(addrs) == 1 {
				host = addrs[0]
			} else if len(addrs) >= 2 {
				host = addrs[0]
				_port, _ := strconv.ParseInt(addrs[1], 10, 0)
				port = int(_port)
			}
		case int:
			port = arg
		}
	} else if len(args) >= 2 {
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = arg
		}
	}

	if len(host) == 0 {
		host = t.ip
		if host == "" {
			host = "0.0.0.0"
		}

	}
	if port == 0 {
		port = 8000
	}
	t.port = port
	addr := host + ":" + strconv.FormatInt(int64(port), 10)

	return addr
}
