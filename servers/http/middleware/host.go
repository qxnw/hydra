package middleware

import (
	x "net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers/http"
)

//Host 处理服务器的主机头
func Host(conf *http.ServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(conf.Hosts) == 0 {
			ctx.Next()
			return
		}
		correct := checkHost(conf.Hosts, ctx)
		if !correct {
			ctx.AbortWithStatus(x.StatusNotAcceptable)
			return
		}
		ctx.Next()
	}
}

func checkHost(hosts []string, ctx *gin.Context) bool {
	chost := ctx.Request.Host
	if i := strings.Index(chost, ":"); i != -1 {
		chost = chost[:i]
	}
	for _, host := range hosts {
		if host == chost {
			return true
		}
	}
	return false

}
