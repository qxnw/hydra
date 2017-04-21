package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/hydra/trace"

	"strings"

	"sync"

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
	"github.com/qxnw/lib4go/metrics"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/sysinfo/memory"
	"github.com/qxnw/lib4go/utility"
	"github.com/spf13/pflag"
)

//Hydra Hydra server
type Hydra struct {
	domain          string
	runMode         string
	tag             string
	mask            string
	system          string
	registry        string
	ip              string
	trace           bool
	collectIndex    int
	registryAddress []string
	*logger.Logger
	watcher conf.ConfWatcher
	servers map[string]*HydraServer
	notify  chan *conf.Updater
	closeCh chan struct{}
	done    bool
	mu      sync.Mutex
}

var (
	SERVER_IS_EXIST     = errors.New("服务已存在")
	SERVER_IS_NOT_EXIST = errors.New("服务不存在")
)

//NewHydra 初始化Hydra服务
func NewHydra() *Hydra {
	return &Hydra{
		servers: make(map[string]*HydraServer),
		Logger:  logger.GetSession("hydra", utility.GetGUID()),
		closeCh: make(chan struct{}, 1),
	}
}

//Install 安装参数
func (h *Hydra) Install() {
	pflag.StringVarP(&h.domain, "domain name", "n", "", "域名称(必须)")
	pflag.StringVarP(&h.registry, "registry center address", "r", "", "注册中心地址(格式：zk://192.168.0.159:2181,192.168.0.158:2181)")
	pflag.StringVarP(&h.mask, "ip mask", "k", "", "ip掩码(本有多个IP时指定，格式:192.168.0)")
	pflag.StringVarP(&h.tag, "server tag", "s", "", "服务器名称(默认为本机IP地址)")
	pflag.BoolVarP(&h.trace, "enable trace", "t", false, "启用项目性能跟踪")
	pflag.BoolVarP(&server.IsDebug, "enable debug", "d", false, "是否启用调试模式")
}
func (h *Hydra) checkFlag() (err error) {
	pflag.Parse()
	if h.domain == "" {
		pflag.Usage()
		return errors.New("domain name 不能为空")
	}
	if h.registry == "" {
		h.runMode = mode_Standalone
	} else {
		h.runMode = mode_cluster
		h.runMode, h.registryAddress, err = getRegistryNames(h.registry)
		if err != nil {
			return fmt.Errorf("集群地址配置有误:%v", err)
		}
	}
	h.ip = net.GetLocalIPAddress(h.mask)
	if h.tag == "" {
		h.tag = h.ip
	}
	return nil

}

//Start 启动服务
func (h *Hydra) Start() (err error) {
	if err = h.checkFlag(); err != nil {
		return
	}
	h.watcher, err = conf.NewWatcher(h.runMode, h.domain, h.tag, h.Logger, h.registryAddress)
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
	go h.collectSys()
	h.Info("启动 hydra server...")

	//启用项目性能跟踪
	if h.trace {
		p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		defer p.Stop()
		go trace.Start(h.Logger)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
LOOP:
	for {
		select {
		case <-interrupt:
			h.Errorf("hydra server(%s) was killed", h.domain)
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
				h.mu.Lock()
				err := h.addServer(u.Conf)
				if err == SERVER_IS_EXIST {
					err = h.changeServer(u.Conf)
				}
				h.mu.Unlock()
			case registry.CHANGE:
				if h.done {
					break LOOP
				}
				h.mu.Lock()
				err := h.changeServer(u.Conf)
				if err == SERVER_IS_NOT_EXIST {
					err = h.addServer(u.Conf)
				}
				h.mu.Unlock()
			case registry.DEL:
				if h.done {
					break LOOP
				}
				h.mu.Lock()
				h.deleteServer(u.Conf)
				h.mu.Unlock()
			}
			break
		}
	}
	for name, srv := range h.servers {
		h.Infof("关闭服务器：%s", name)
		srv.Shutdown()
	}
	h.closeCh <- struct{}{}
	return nil
}
func (h *Hydra) addServer(cnf conf.Conf) error {
	name := cnf.String("name")
	if _, ok := h.servers[name]; ok {
		return SERVER_IS_EXIST
	}
	h.Logger.Infof("启动服务器:%s", name)
	srv := NewHydraServer(h.domain, h.runMode, h.registry, h.registryAddress)
	err := srv.Start(cnf)
	if err != nil {
		h.Error(err)
		return err
	}
	h.servers[name] = srv
	return nil
}
func (h *Hydra) changeServer(cnf conf.Conf) error {
	name := cnf.String("name")
	h.Logger.Infof("配置发生变化:%s", name)
	srv, ok := h.servers[name]
	if !ok {
		return SERVER_IS_NOT_EXIST
	}

	err := srv.Notify(cnf)
	if err != nil {
		h.Errorf("server:%s(err:%v)", name, err)
		h.deleteServer(cnf)
	}
	return err
}
func (h *Hydra) deleteServer(cnf conf.Conf) {
	name := cnf.String("name")
	h.Logger.Infof("服务器被删除:%s", name)
	if srv, ok := h.servers[name]; ok {
		srv.Shutdown()
		delete(h.servers, name)
	}
}

func (h *Hydra) Close() {
	h.done = true
	<-h.closeCh
	time.Sleep(time.Millisecond * 100)
	if h.watcher != nil {
		h.watcher.Close()
	}
	logger.Close()
	close(h.closeCh)
}
func getRegistryNames(address string) (clusterName string, raddr []string, err error) {
	addr := strings.SplitN(address, "://", 2)
	if len(addr) != 2 {
		return "", nil, fmt.Errorf("%s错误，必须包含://", addr)
	}
	if len(addr[0]) == 0 {
		return "", nil, fmt.Errorf("%s错误，协议名不能为空", addr)
	}
	if len(addr[1]) == 0 {
		return "", nil, fmt.Errorf("%s错误，地址不能为空", addr)
	}
	clusterName = addr[0]
	raddr = strings.Split(addr[1], ",")
	return
}
func (h *Hydra) collectSys() {
	for {
		select {
		case <-time.After(time.Second * 5):
			if h.done {
				return
			}
			h.collectIndex = (h.collectIndex + 1) % 5
			if h.collectIndex != 0 {
				continue
			}
			cpuUsed := metrics.GetOrRegisterGaugeFloat64(metrics.MakeName("hydra.server.cpu", metrics.GAUGE, "server", h.ip), metrics.DefaultRegistry)   //响应时长
			memUsed := metrics.GetOrRegisterGaugeFloat64(metrics.MakeName("hydra.server.mem", metrics.GAUGE, "server", h.ip), metrics.DefaultRegistry)   //响应时长
			diskUsed := metrics.GetOrRegisterGaugeFloat64(metrics.MakeName("hydra.server.disk", metrics.GAUGE, "server", h.ip), metrics.DefaultRegistry) //响应时长

			u := cpu.GetInfo()
			cpuUsed.Update(u.UsedPercent)
			mm := memory.GetInfo()
			memUsed.Update(mm.UsedPercent)
			dsk := disk.GetInfo()
			diskUsed.Update(dsk.UsedPercent)
		}
	}
}
