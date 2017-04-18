package trace

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"

	"github.com/qxnw/lib4go/logger"
)

//Start 启用项目跟踪
func Start(log *logger.Logger) error {
	f, err := os.Create("trace.out")
	if err != nil {
		return err
	}
	defer f.Close()
	err = trace.Start(f)
	if err != nil {
		return err
	}
	defer trace.Stop()
	addr := "localhost:19999"
	log.Info("启用项目跟踪:http://localhost:19999/debug/pprof/")
	return http.ListenAndServe(addr, nil)
}
