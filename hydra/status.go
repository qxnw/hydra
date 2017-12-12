package hydra

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/sysinfo/memory"
)

type HydraServer struct {
	Servers   []*ServerInfo `json:"servers"`
	AppMemory uint64        `json:"app_memory"`
	CPUUsed   string        `json:"cpu_used_precent"`
	MemUsed   string        `json:"mem_used_precent"`
	DiskUsed  string        `json:"disk_used_precent"`
}

//ServerInfo 服务信息
type ServerInfo struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Start    int64    `json:"start"`
	Services []string `json:"services,omitempty"`
}

var statusLocalPort = []int{10160, 10161, 10162, 10163, 10164, 10165, 10166, 10167}

//StartStatusServer 启动状态服务器
func (h *Hydra) StartStatusServer(domain string) (err error) {
	h.ws = api.NewAPI(domain, "status.server")
	h.ws.Route("GET", "/server/query", func(c *api.Context) {
		h.queryServerStatus(c)
	})
	h.ws.Route("GET", "/server/update/:version", func(c *api.Context) {
		h.update(c)
	})

	go func() {
		err = h.ws.Run(net.GetAvailablePort(statusLocalPort))
		if err != nil {
			h.Error("ws:", err)
		}
		return
	}()
	return nil
}

//--------------------------------------服务器相关操作----------------------------------------------------
func (h *Hydra) queryServerStatus(c *api.Context) {
	hydraServer := &HydraServer{}
	hydraServer.AppMemory = memory.GetAPPMemory()
	hydraServer.CPUUsed = fmt.Sprintf("%.2f", cpu.GetInfo(time.Millisecond*200).UsedPercent)
	hydraServer.MemUsed = fmt.Sprintf("%.2f", memory.GetInfo().UsedPercent)
	hydraServer.DiskUsed = fmt.Sprintf("%.2f", disk.GetInfo().UsedPercent)
	hydraServer.Servers = make([]*ServerInfo, 0, len(h.servers))
	for _, v := range h.servers {
		hydraServer.Servers = append(hydraServer.Servers, &ServerInfo{
			Name:     fmt.Sprintf("%s/servers/%s/%s", v.domain, v.serverName, v.serverType),
			Start:    v.runTime.Unix(),
			Address:  v.address,
			Services: v.localServices,
		})
	}
	c.Result = &api.StatusResult{Code: 200, Result: hydraServer, Type: api.JsonResponse}
	return
}

func (h *Hydra) update(c *api.Context) {
	h.Info("启动软件更新")
	version := c.Param("version")
	pkg, err := h.getPackage(version)
	if err != nil {
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	if version != pkg.Version {
		err = fmt.Errorf("安装包配置的版本号有误:%s(%s)", version, pkg.Version)
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	err = h.updateNow(pkg.URL)
	if err != nil {
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	err = h.restartHydra()
	if err != nil {
		c.Result = &api.StatusResult{Code: 500, Result: err.Error(), Type: 0}
		return
	}
	c.Result = &api.StatusResult{Code: 200, Result: "SUCCESS", Type: 0}
	return
}
