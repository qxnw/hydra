package main

import (
	"fmt"

	"github.com/qxnw/hydra/registry/conf"
	_ "github.com/qxnw/hydra/registry/conf/json"
	_ "github.com/qxnw/hydra/registry/conf/registry"
	"github.com/qxnw/lib4go/net"
	"github.com/spf13/pflag"
)

type Hydra struct {
	domain string
	mode   string
	mask   string
}

//Install 安装参数
func (h *Hydra) Install() {
	pflag.StringVarP(&h.domain, "domain", "d", "", "域名称")
	pflag.StringVarP(&h.mode, "mode", "m", "", "配置文件类型")
	pflag.StringVarP(&h.mask, "mask", "k", "", "配置文件类型")
}

//Start 启动服务
func (h *Hydra) Start() {
	pflag.Parse()
	ip := net.GetLocalIPAddress(h.mask)
	watcher, err := conf.NewWatcher(h.mode, h.domain, ip)
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case u := <-watcher.Notify():
			break
		}
	}
}
