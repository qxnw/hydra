package hydra

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"

	"github.com/qxnw/lib4go/logger"
)

//StartTraceServer 启用性能跟踪(协程，内存，堵塞等)
func StartTraceServer(log *logger.Logger) error {
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
	addr := "0.0.0.0:19999"
	log.Info("启用项目跟踪:http://0.0.0.0:19999/debug/pprof/")
	return http.ListenAndServe(addr, nil)
}
