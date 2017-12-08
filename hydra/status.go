package hydra

import (
	"encoding/json"

	"strings"

	"fmt"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/jsons"
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
		h.queryServerStatus(c)
	})
	ws.Route("GET", "/sys/server/query/:name", func(c *api.Context) {
		h.queryServerStatusByName(c)
	})
	ws.Route("POST", "/sys/server/update/:version", func(c *api.Context) {
		h.update(c)
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

//--------------------------------------服务器相关操作----------------------------------------------------
func (h *Hydra) queryServerStatus(c *api.Context) {
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
}

func (h *Hydra) queryServerStatusByName(c *api.Context) {
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
}
func (h *Hydra) update(c *api.Context) {
	version := c.Param("version")
	setting, err := h.getSetting(version)
	if err != nil {
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	if version != setting.Version {
		err = fmt.Errorf("更新报有误，配置的版本号不一致")
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	err = h.updateNow(setting.URL)
	if err != nil {
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
}
