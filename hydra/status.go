package hydra

import (
	"fmt"
	"time"

	"github.com/qxnw/hydra/context"
	xhttp "github.com/qxnw/hydra/servers/http"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/sysinfo/memory"
)

type HydraServer struct {
	Version   string        `json:"version"`
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
	conf := conf.NewConf(domain, "health-checker", "api", h.tag, "", "", 3)
	h.healthChecker, err = xhttp.NewApiServer(conf, nil, xhttp.WithIP(conf.IP), xhttp.WithLogger(h.Logger))
	if err != nil {
		h.Error("health-checker:", err)
		return err
	}
	routers := xhttp.GetRouters()
	routers.Route("GET", "/server/query", func(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
		return h.queryServerStatus(name, engine, service, ctx)
	})
	routers.Route("GET", "/server/update/:systemName/:version", func(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
		return h.update(name, engine, service, ctx)
	})
	h.healthChecker.SetRouters(routers.Get())

	port := net.GetAvailablePort(statusLocalPort)
	err = h.healthChecker.Run(port)
	if err != nil {
		h.Error("health-checker:", err)
		return err
	}
	h.Infof("启动成功:health-checker.api(addr:%s)", h.healthChecker.GetAddress())
	return nil
}

//--------------------------------------服务器相关操作----------------------------------------------------
func (h *Hydra) queryServerStatus(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
	hydraServer := &HydraServer{}
	hydraServer.Version = Version
	hydraServer.AppMemory = memory.GetAPPMemory()
	hydraServer.CPUUsed = fmt.Sprintf("%.2f", cpu.GetInfo(time.Millisecond*200).UsedPercent)
	hydraServer.MemUsed = fmt.Sprintf("%.2f", memory.GetInfo().UsedPercent)
	hydraServer.DiskUsed = fmt.Sprintf("%.2f", disk.GetInfo().UsedPercent)
	hydraServer.Servers = make([]*ServerInfo, 0, len(h.servers))
	for _, v := range h.servers {
		hydraServer.Servers = append(hydraServer.Servers, &ServerInfo{
			Name:     fmt.Sprintf("/%s/servers/%s/%s", v.domain, v.serverName, v.serverType),
			Start:    v.runTime.Unix(),
			Address:  v.address,
			Services: v.localServices,
		})
	}
	response := context.GetObjectResponse()
	response.SetContent(200, hydraServer)
	return response, nil
}

func (h *Hydra) update(name string, engine string, service string, ctx *context.Context) (rs context.Response, err error) {
	h.Info("启动软件更新")
	response := context.GetStandardResponse()
	version := ctx.Request.Param.GetString("version")
	systemName := ctx.Request.Param.GetString("systemName")
	if version == Version {
		response.SetContent(204, "无需更新")
		return response, nil
	}
	pkg, err := h.getPackage(systemName, version)
	if err != nil {
		h.Error(err)
		response.SetContent(500, err)
		return response, err
	}
	if version != pkg.Version {
		err = fmt.Errorf("安装包配置的版本号有误:%s(%s)", version, pkg.Version)
		h.Error(err)
		response.SetContent(500, err)
		return response, err
	}
	err = h.updateNow(pkg.URL, pkg.CRC32)
	if err != nil {
		h.Error(err)
		response.SetContent(500, err)
		return response, err
	}
	err = h.restartHydra()
	if err != nil {
		h.Error(err)
		response.SetContent(500, err)
		return response, err
	}
	response.SetContent(200, "success")
	return response, nil
}
