package web

import (
	"net/http"
	"strings"
)

func Host() HandlerFunc {
	return func(ctx *Context) {
		if len(ctx.tan.hostNames) == 0 {
			ctx.Next()
			return
		}
		host := getHost(ctx.Req())
		exist := false
		for _, h := range ctx.tan.hostNames {
			if h == host {
				exist = true
				break
			}
		}
		if !exist {
			ctx.Abort(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable))
			return
		}
		ctx.Next()
	}
}

func getHost(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.Host
	}
	host := r.Host
	if i := strings.Index(host, ":"); i != -1 {
		host = host[:i]
	}
	return host

}
