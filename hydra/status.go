package hydra

import (
	"encoding/json"

	"strings"

	"fmt"

	"github.com/qxnw/hydra/server/web"
	"github.com/qxnw/lib4go/jsons"
)

type ServerStatus struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Start    int64    `json:"start"`
	Services []string `json:"srvs,omitempty"`
}

func (h *Hydra) StartStatusServer() (err error) {
	ws := web.New("status.server")
	ws.Route("GET", "/sys/server/query", func(c *web.Context) {
		status := make([]ServerStatus, 0, len(h.servers))
		for _, v := range h.servers {
			status = append(status, ServerStatus{Name: fmt.Sprintf("%s/%s/%s", v.domain, v.serverName, v.serverType), Start: v.runTime.Unix(), Address: v.address})
		}
		buf, err := jsons.Marshal(status)
		if err != nil {
			c.Result = &web.StatusResult{Code: 500, Result: "Internal Server Error(工作引擎发生异常)", Type: 0}
			return
		}
		c.Result = &web.StatusResult{Code: 200, Result: json.RawMessage(buf), Type: web.JsonResponse}
		return
	})
	ws.Route("GET", "/sys/server/query/:name", func(c *web.Context) {
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
			c.Result = &web.StatusResult{Code: 500, Result: "Internal Server Error(工作引擎发生异常)", Type: 0}
			return
		}
		c.Result = &web.StatusResult{Code: 200, Result: json.RawMessage(buf), Type: web.JsonResponse}
		return
	})

	go func() error {
		err = ws.Run(":10160")
		if err != nil {
			return err
		}
		return nil
	}()
	return nil
}
