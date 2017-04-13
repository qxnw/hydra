package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qxnw/hydra/conf"
	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	"github.com/qxnw/hydra/registry"
	_ "github.com/qxnw/hydra/server/cron"
	_ "github.com/qxnw/hydra/server/mq"
	_ "github.com/qxnw/hydra/server/rpc"
	_ "github.com/qxnw/hydra/server/web"
	_ "github.com/qxnw/hydra/service/discovery"
	_ "github.com/qxnw/hydra/service/register"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/utility"
	"github.com/spf13/pflag"
)

//Hydra Hydra server
type Hydra struct {
	domain   string
	runMode  string
	tag      string
	mask     string
	system   string
	registry string
	*logger.Logger
	watcher conf.ConfWatcher
	servers map[string]*HydraServer
	notify  chan *conf.Updater
	done    bool
}

//NewHydra 初始化Hydra服务
func NewHydra() *Hydra {
	return &Hydra{
		servers: make(map[string]*HydraServer),
		Logger:  logger.GetSession("hydra", utility.GetGUID()),
	}
}

//Install 安装参数
func (h *Hydra) Install() {
	pflag.StringVarP(&h.domain, "domain", "d", "", "域名称")
	pflag.StringVarP(&h.runMode, "run mode", "m", "standalone", "运行模式(standalone,cluster)")
	pflag.StringVarP(&h.registry, "registry", "r", "", "服务注册中心地址(运行模式为cluster时必须填写)")
	pflag.StringVarP(&h.mask, "ip mask", "k", "", "ip掩码(默认为本机第一个有效IP)")
	pflag.StringVarP(&h.tag, "server tag", "t", "", "服务器标识(默认为本机IP地址)")
}
func (h *Hydra) checkFlag() bool {
	pflag.Parse()
	if h.domain == "" || h.runMode == "" {
		pflag.Usage()
		return false
	}
	if h.runMode == mode_cluster && h.registry == "" {
		pflag.Usage()
		return false
	}
	if h.tag == "" {
		h.tag = net.GetLocalIPAddress(h.mask)
	}
	return true

}

//Start 启动服务
func (h *Hydra) Start() (err error) {
	if !h.checkFlag() {
		return errors.New("输入参数为空")
	}
	h.watcher, err = conf.NewWatcher(h.runMode, h.domain, h.tag)
	if err != nil {
		h.Error(fmt.Sprintf("watcher初始化失败 run mode:%s,domain:%s(err:%v)", h.runMode, h.domain, err))
		return
	}
	h.notify = h.watcher.Notify()
	err = h.watcher.Start()
	if err != nil {
		h.Error(fmt.Sprintf("watcher启用失败 run mode:%s,domain:%s(err:%v)", h.runMode, h.domain, err))
		return
	}
	go h.loopCheckNotify()
	h.Info("启动 hydra server...")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
LOOP:
	for {
		select {
		case <-interrupt:
			h.Errorf("%s was killed", h.domain)
			h.done = true
			break LOOP
		}
	}
	return nil
}

func (h *Hydra) loopCheckNotify() (err error) {
LOOP:
	for {
		select {
		case <-time.After(time.Second):
			if h.done {
				break LOOP
			}
		case u := <-h.notify:
			switch u.Op {
			case registry.ADD:
				if h.done {
					break LOOP
				}
				name := u.Conf.String("name")
				h.Logger.Infof("start server:%s", name)
				srv := NewHydraServer(h.domain, h.runMode, h.registry)
				err = srv.Start(u.Conf)
				if err != nil {
					h.Error(err)
				}
				h.servers[name] = srv
			case registry.CHANGE:
				if h.done {
					break LOOP
				}
				name := u.Conf.String("name")
				h.Logger.Info("conf changed:%s", name)
				if srv, ok := h.servers[name]; ok {
					err = srv.Notify(u.Conf)
					if err != nil {
						h.Errorf("配置更新失败 server:%s(err:%v)", name, err)
						return
					}
				}
			case registry.DEL:
				if h.done {
					break LOOP
				}
				name := u.Conf.String("name")
				h.Logger.Info("close server:%s", name)
				if srv, ok := h.servers[name]; ok {
					srv.Shutdown()
					h.Infof("关闭服务器：%s", name)
					delete(h.servers, name)
				}
			}
			break
		}
	}
	for name, srv := range h.servers {
		srv.Shutdown()
		h.Infof("关闭服务器：%s", name)
	}
	return nil
}
