package hydra

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/qxnw/hydra/pprof"
	"github.com/qxnw/hydra/server"

	"sync"

	"strings"

	"runtime/debug"

	"github.com/qxnw/hydra/conf"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"

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
	trace                  string
	collectIndex           int
	currentRegistryAddress []string
	crossRegistryAddress   []string
	*logger.Logger
	watcher      conf.ConfWatcher
	servers      map[string]*HydraServer
	notify       chan *conf.Updater
	closedNotify chan struct{}
	closeChan    chan struct{}
	done         bool
	mu           sync.Mutex
}

var (
	SERVER_IS_EXIST     = errors.New("服务已存在")
	SERVER_IS_NOT_EXIST = errors.New("服务不存在")
)

//NewHydra 初始化Hydra服务
func NewHydra() *Hydra {
	return &Hydra{
		servers:      make(map[string]*HydraServer),
		Logger:       logger.GetSession("hydra", logger.CreateSession()),
		closedNotify: make(chan struct{}, 1),
		closeChan:    make(chan struct{}),
	}
}

//Install 安装参数
func (h *Hydra) Install() {
	//pflag.StringVarP(&h.domain, "domain name", "n", "", "域名称(必须)")
	pflag.StringVarP(&h.currentRegistry, "registry center address", "r", "", "注册中心地址(格式：zk://192.168.0.159:2181,192.168.0.158:2181)")
	pflag.StringVarP(&h.mask, "ip mask", "i", "", "ip掩码(本有多个IP时指定，格式:192.168.0)")
	pflag.StringVarP(&h.tag, "server tag", "t", "", "服务器名称(默认为本机IP地址)")
	pflag.StringVarP(&h.trace, "enable trace", "p", "", "启用项目性能跟踪cpu/mem/block/mutex/server")
	pflag.BoolVarP(&server.IsDebug, "enable debug", "d", false, "是否启用调试模式")
	pflag.StringVarP(&h.crossRegistry, "cross  registry  center address", "c", "", "跨域注册中心地址")
}
func (h *Hydra) checkFlag() (err error) {
	pflag.Parse()
	if len(os.Args) < 2 {
		return errors.New("第一个参数必须为域名称")
	}
	engine.IsDebug = server.IsDebug
	h.domain = os.Args[1]
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
		h.Error(err)
		return
	}

	if err = h.StartStatusServer(h.domain); err != nil {
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
		h.Errorf("watcher启用失败 run mode:%s,domain:%s(err:%v)", h.runMode, h.domain, err)
		return
	}
	go h.loopCheckNotify()
	go h.freeMemory()
	//go h.collectSys()
	h.Infof("启动 hydra server(%s)...", h.tag)
	//启用项目性能跟踪
	switch h.trace {
	case "cpu":
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	case "web":
		go pprof.StartTraceServer(h.Logger)
	default:
		h.Logger.Info("未启用项目 跟踪")
	}

	interrupt := make(chan os.Signal, 1)
	//signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM) //9:kill/SIGKILL,15:SIGTEM,20,SIGTOP 2:interrupt/syscall.SIGINT
LOOP:
	for {
		select {
		case <-interrupt:
			h.Warnf("hydra server(%s) was killed", h.domain)
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
		case <-h.closeChan:
			break LOOP
		case u := <-h.notify:
			if h.done {
				break LOOP
			}
			u.Conf.Set("tag", h.tag)
			switch u.Op {
			case registry.ADD:
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
	h.closedNotify <- struct{}{}
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
	h.Logger.Infof("启动成功:%s(addr:%s,srvs:%d)", name, srv.address, len(srv.localServices))
	return nil
}
func (h *Hydra) changeServer(cnf conf.Conf) error {
	name := h.getServerName(cnf)
	h.Logger.Warnf("配置发生变化:%s", name)
	srv, ok := h.servers[name]
	if !ok {
		return SERVER_IS_NOT_EXIST
	}
	if srv.EngineConfChanged(cnf.String("extModes")) {
		h.deleteServer(cnf)
		srv.Shutdown()
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
	close(h.closeChan)
	h.done = true
	registry.Close()
	if len(h.servers) > 0 {
		select {
		case <-h.closedNotify:
		case <-time.After(time.Second * 3):
		}

	}
	time.Sleep(time.Millisecond * 100)
	if h.watcher != nil {
		h.watcher.Close()
	}
	logger.Close()
	close(h.closedNotify)

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
