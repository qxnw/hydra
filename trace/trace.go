package trace

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/qxnw/lib4go/logger"
)

//Start 启用项目跟踪
func Start(log *logger.Logger) error {
	addr := "localhost:19999"
	log.Info("启用项目跟踪:http://localhost:19999/debug/pprof/")
	return http.ListenAndServe(addr, nil)
}
