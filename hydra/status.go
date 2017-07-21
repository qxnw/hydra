package hydra

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime/trace"

	"strings"

	"fmt"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/net"
)

//StatusServer 状态服务器，用于查询hydra启动了哪些服务器及提供了哪些服务
type StatusServer struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Start    int64    `json:"start"`
	Services []string `json:"srvs,omitempty"`
}

var statusLocalPort = []int{10160, 10161, 10162, 10163, 10164, 10165, 10166, 10167}

//StartStatusServer 启动状态服务器
func (h *Hydra) StartStatusServer(domain string) (err error) {
	ws := api.NewAPI(domain, "status.server")
	ws.Route("GET", "/sys/server/query", func(c *api.Context) {
		status := make([]StatusServer, 0, len(h.servers))
		for _, v := range h.servers {
			status = append(status, StatusServer{Name: fmt.Sprintf("%s/%s/%s", v.domain, v.serverName, v.serverType), Start: v.runTime.Unix(), Address: v.address})
		}
		buf, err := jsons.Marshal(status)
		if err != nil {
			c.Result = &api.StatusResult{Code: 500, Result: "Internal Server Error(工作引擎发生异常)", Type: 0}
			return
		}
		c.Result = &api.StatusResult{Code: 200, Result: json.RawMessage(buf), Type: api.JsonResponse}
		return
	})
	ws.Route("GET", "/sys/server/query/:name", func(c *api.Context) {
		server := c.Param("name")
		status := make([]StatusServer, 0, 1)
		for _, v := range h.servers {
			if strings.Contains(v.serverName, server) {
				status = append(status, StatusServer{
					Name:     fmt.Sprintf("%s/%s/%s", v.domain, v.serverName, v.serverType),
					Start:    v.runTime.Unix(),
					Address:  v.address,
					Services: v.localServices,
				})
			}
		}
		buf, err := jsons.Marshal(status)
		if err != nil {
			c.Result = &api.StatusResult{Code: 500, Result: "Internal Server Error(工作引擎发生异常)", Type: 0}
			return
		}
		c.Result = &api.StatusResult{Code: 200, Result: json.RawMessage(buf), Type: api.JsonResponse}
		return
	})

	go func() error {
		err = ws.Run(net.GetAvailablePort(statusLocalPort))
		if err != nil {
			return err
		}
		return nil
	}()
	return nil
}

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
