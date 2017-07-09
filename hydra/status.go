package hydra

import (
	"encoding/json"

	"strings"

	"fmt"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/jsons"
)

type ServerStatus struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Start    int64    `json:"start"`
	Services []string `json:"srvs,omitempty"`
}

func (h *Hydra) StartStatusServer(domain string) (err error) {
	ws := api.New(domain, "status.server")
	ws.Route("GET", "/sys/server/query", func(c *api.Context) {
		status := make([]ServerStatus, 0, len(h.servers))
		for _, v := range h.servers {
			status = append(status, ServerStatus{Name: fmt.Sprintf("%s/%s/%s", v.domain, v.serverName, v.serverType), Start: v.runTime.Unix(), Address: v.address})
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
		status := make([]ServerStatus, 0, 1)
		for _, v := range h.servers {
			if strings.Contains(v.serverName, server) {
				status = append(status, ServerStatus{
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
		err = ws.Run(":10161")
		if err != nil {
			return err
		}
		return nil
	}()
	return nil
}
