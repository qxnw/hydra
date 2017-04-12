package main

import (
	"fmt"

	"github.com/qxnw/hydra/registry/conf"
	_ "github.com/qxnw/hydra/registry/conf/json"
	_ "github.com/qxnw/hydra/registry/conf/registry"
	"github.com/qxnw/lib4go/net"
	"github.com/spf13/pflag"
)

func main() {
	var domain string
	var mode string
	var mask string
	pflag.StringVarP(&domain, "domain", "d", "", "域名称")
	pflag.StringVarP(&mode, "mode", "m", "", "配置文件类型")
	pflag.StringVarP(&mask, "mask", "k", "", "配置文件类型")
	pflag.Parse()
	tag := net.GetLocalIPAddress(mask)
	watcher, err := conf.NewWatcher(mode, domain, tag)
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case u := <-watcher.Notify():
			fmt.Println(u.Op, u.Conf)

		}
	}
}
