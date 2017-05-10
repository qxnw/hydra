package hydra

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

	"sync"

	"strings"

	"runtime/debug"

	"github.com/qxnw/hydra/conf"
	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	"github.com/qxnw/hydra/engine"
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
	"github.com/qxnw/lib4go/transform"
	"github.com/spf13/pflag"
)

//Hydra Hydra server
type Hydra struct {
	domain                 string
	runMode                string
	tag                    string
	mask                   string
	system                 string
	currentRegistry        string
	crossRegistry          string
	ip                     string
	baseData               *transform.Transform
	trace                  bool
	collectIndex           int
	currentRegistryAddress []string
	crossRegistryAddress   []string
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
		Logger:  logger.GetSession("hydra", logger.CreateSession()),
		closeCh: make(chan struct{}, 1),
	}
}

//Install 安装参数
func (h *Hydra) Install() {
	pflag.StringVarP(&h.domain, "domain name", "n", "", "域名称(必须)")
	pflag.StringVarP(&h.currentRegistry, "registry center address", "r", "", "注册中心地址(格式：zk://192.168.0.159:2181,192.168.0.158:2181)")
	pflag.StringVarP(&h.mask, "ip mask", "i", "", "ip掩码(本有多个IP时指定，格式:192.168.0)")
	pflag.StringVarP(&h.tag, "server tag", "t", "", "服务器名称(默认为本机IP地址)")
	pflag.BoolVarP(&h.trace, "enable trace", "p", false, "启用项目性能跟踪")
	pflag.BoolVarP(&server.IsDebug, "enable debug", "d", false, "是否启用调试模式")
	pflag.StringVarP(&h.crossRegistry, "cross  registry  center address", "c", "", "跨域注册中心地址")
}
func (h *Hydra) checkFlag() (err error) {
	pflag.Parse()
	engine.IsDebug = server.IsDebug
	if h.domain == "" {
		pflag.Usage()
		return errors.New("domain name 不能为空")
	}
	if h.currentRegistry == "" {
		h.runMode = mode_Standalone
		h.currentRegistryAddress = []string{"localhost"}
		h.currentRegistry = fmt.Sprintf("%s://%s", h.runMode, strings.Join(h.currentRegistryAddress, ","))
	} else {
		h.runMode = mode_cluster
		h.runMode, h.currentRegistryAddress, err = registry.ResolveAddress(h.currentRegistry)
		if err != nil {
			return fmt.Errorf("集群地址配置有误:%v", err)
		}
	}
	if h.crossRegistry != "" {
		if strings.Contains(h.crossRegistry, "//") {
			return fmt.Errorf("跨域注册中心地址不能指定协议信息:%s(err:%v)", h.crossRegistry, err)
		}
		h.crossRegistryAddress = strings.Split(h.crossRegistry, ",")
	}
	h.ip = net.GetLocalIPAddress(h.mask)
	if h.tag == "" {
		h.tag = h.ip
	}
	h.baseData = transform.NewMap(map[string]string{
		"ip": h.ip,
	})
	h.tag = h.baseData.Translate(h.tag)
	return nil

}

//Start 启动服务
func (h *Hydra) Start() (err error) {
	if err = h.checkFlag(); err != nil {
		return
	}

	if err = h.StartStatusServer(); err != nil {
		return
	}
	h.watcher, err = conf.NewWatcher(h.runMode, h.domain, h.tag, h.Logger, h.currentRegistryAddress)
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
	go h.freeMemory()
	//go h.collectSys()
	h.Infof("启动 hydra server(%s)...", h.tag)
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
			u.Conf.Set("tag", h.tag)
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
				if err != nil {
					h.Error(err)
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
				if err != nil {
					h.Error(err)
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
		h.Warnf("关闭服务器：%s", name)
		srv.Shutdown()
	}
	h.closeCh <- struct{}{}
	return nil
}
func (h *Hydra) getServerName(cnf conf.Conf) string {
	return fmt.Sprintf("%s_%s", cnf.String("name"), cnf.String("type"))
}
func (h *Hydra) addServer(cnf conf.Conf) error {
	name := h.getServerName(cnf)
	if _, ok := h.servers[name]; ok {
		return SERVER_IS_EXIST
	}
	srv := NewHydraServer(h.domain, h.runMode, h.currentRegistry, h.Logger, h.currentRegistryAddress, h.crossRegistryAddress)
	err := srv.Start(cnf)
	if err != nil {
		return err
	}
	h.servers[name] = srv
	h.Logger.Infof("启动服务器:%s(%s)", name, srv.address)
	return nil
}
func (h *Hydra) changeServer(cnf conf.Conf) error {
	name := h.getServerName(cnf)
	h.Logger.Warnf("配置发生变化:%s", name)
	srv, ok := h.servers[name]
	if !ok {
		return SERVER_IS_NOT_EXIST
	}

	err := srv.Notify(cnf)
	if err != nil || srv.GetStatus() == server.ST_STOP {
		h.deleteServer(cnf)
		srv.Shutdown()
	}
	return err
}
func (h *Hydra) deleteServer(cnf conf.Conf) {
	h.deleteServerByID(h.getServerName(cnf))
}
func (h *Hydra) deleteServerByID(id string) {
	h.Logger.Warnf("服务器被删除:%s", id)
	if srv, ok := h.servers[id]; ok {
		srv.Shutdown()
		delete(h.servers, id)
	}
}

//Close 关闭服务器
func (h *Hydra) Close() {
	h.done = true
	registry.Close()
	select {
	case <-h.closeCh:
	case <-time.After(time.Second):
	}
	time.Sleep(time.Millisecond * 100)
	if h.watcher != nil {
		h.watcher.Close()
	}
	logger.Close()
	close(h.closeCh)

}

func (h *Hydra) freeMemory() {
	for {
		select {
		case <-time.After(time.Second * 120):
			debug.FreeOSMemory()
		}
	}
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
