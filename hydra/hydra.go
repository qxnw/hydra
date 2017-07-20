package hydra

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/qxnw/hydra/server"

	"sync"

	"strings"

	"runtime/debug"

	"github.com/qxnw/hydra/conf"

	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/hydra/registry"

	log "github.com/qxnw/hydra/logger"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/net"
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
	watcher      conf.Watcher
	servers      map[string]*Server
	notify       chan *conf.Updater
	closedNotify chan struct{}
	closeChan    chan struct{}
	done         bool
	mu           sync.Mutex
	rpcLogger    bool
}

//NewHydra 初始化Hydra服务
func NewHydra() *Hydra {
	h := &Hydra{
		servers:      make(map[string]*Server),
		Logger:       logger.GetSession("hydra", logger.CreateSession()),
		closedNotify: make(chan struct{}, 1),
		closeChan:    make(chan struct{}),
	}
	pflag.StringVarP(&h.currentRegistry, "registry center address", "r", "", "注册中心地址(格式：zk://192.168.0.159:2181,192.168.0.158:2181)")
	pflag.StringVarP(&h.mask, "ip mask", "i", "", "ip掩码(本有多个IP时指定，格式:192.168.0)")
	pflag.StringVarP(&h.tag, "server tag", "t", "", "服务器名称(默认为本机IP地址)")
	pflag.StringVarP(&h.trace, "enable trace", "p", "", "启用项目性能跟踪cpu/mem/block/mutex/server")
	pflag.BoolVarP(&server.IsDebug, "enable debug", "d", false, "是否启用调试模式")
	pflag.StringVarP(&h.crossRegistry, "cross  registry  center address", "c", "", "跨域注册中心地址")
	pflag.BoolVarP(&h.rpcLogger, "use rpc logger", "g", false, "使用RPC远程记录日志")
	return h
}

func (h *Hydra) checkFlag() (err error) {
	pflag.Parse()
	if len(os.Args) < 2 {
		return errors.New("未指定域名称")
	}
	engine.IsDebug = server.IsDebug
	h.domain = os.Args[1]
	if h.currentRegistry == "" {
		h.runMode = modeStandalone
		h.currentRegistryAddress = []string{"localhost"}
		h.currentRegistry = fmt.Sprintf("%s://%s", h.runMode, strings.Join(h.currentRegistryAddress, ","))
	} else {
		h.runMode = modeCluster
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

//Start 启动服务器
func (h *Hydra) Start() (err error) {
	defer h.recovery()
	if err = h.checkFlag(); err != nil {
		h.Error(err)
		return
	}
	//检查是否配置RPC日志服务
	if h.rpcLogger && h.runMode != modeStandalone {
		err = log.ConfigRPCLogger(h.domain, h.currentRegistry, h.Logger)
		if err != nil {
			h.Errorf("无法启用RPC日志:%v", err)
			return
		}
		h.Info("hydra:启用RPC日志")
	}
	//启动服务器状态查询服务
	if err = h.StartStatusServer(h.domain); err != nil {
		return
	}

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
		go StartTraceServer(h.Logger)
	default:
	}

	//启动服务器配置监控
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
	go h.loopRecvNotify()
	go h.freeMemory()
	h.Infof("启动成功 hydra server(%s,%s)...", h.tag, h.runMode)

	//监听操作系统事件ctrl+c, kill
	interrupt := make(chan os.Signal, 1)
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

//循环接收服务器配置变化通知
func (h *Hydra) loopRecvNotify() (err error) {
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
				if err == errServerIsExist {
					err = h.changeServer(u.Conf)
				}
				if err != nil {
					h.Error(err)
				}
				h.mu.Unlock()
			case registry.CHANGE:
				h.mu.Lock()
				err := h.changeServer(u.Conf)
				if err == errServerIsNotExist {
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

//获取服务器名称
func (h *Hydra) getServerName(cnf conf.Conf) string {
	return fmt.Sprintf("%s_%s", cnf.String("name"), cnf.String("type"))
}

//添加新服务器
func (h *Hydra) addServer(cnf conf.Conf) error {
	name := h.getServerName(cnf)
	if _, ok := h.servers[name]; ok {
		return errServerIsExist
	}
	srv := NewHydraServer(h.domain, h.runMode, h.currentRegistry, h.currentRegistryAddress, h.crossRegistryAddress, h.Logger)
	err := srv.Start(cnf)
	if err != nil {
		return err
	}
	h.servers[name] = srv
	h.Logger.Infof("启动成功:%s(addr:%s,srvs:%d)", name, srv.address, len(srv.localServices))
	return nil
}

//服务器配置变化
func (h *Hydra) changeServer(cnf conf.Conf) error {
	name := h.getServerName(cnf)
	h.Logger.Warnf("配置发生变化:%s", name)
	srv, ok := h.servers[name]
	if !ok {
		return errServerIsNotExist
	}
	if srv.EngineHasChange(cnf.String("extModes")) {
		h.deleteServer(cnf)
		srv.Shutdown()
		return errServerIsNotExist
	}

	err := srv.Notify(cnf)
	if err != nil || srv.GetStatus() == server.ST_STOP {
		h.deleteServer(cnf)
		srv.Shutdown()
	}
	return err
}

//根据配置删除服务器
func (h *Hydra) deleteServer(cnf conf.Conf) {
	h.deleteServerByID(h.getServerName(cnf))
}

//根据ID删除服务器
func (h *Hydra) deleteServerByID(id string) {
	h.Logger.Warnf("服务器被删除:%s", id)
	if srv, ok := h.servers[id]; ok {
		srv.Shutdown()
		delete(h.servers, id)
	}
}

//Close 关闭hydra服务器
func (h *Hydra) Close() {
	h.done = true
	close(h.closeChan)
	if len(h.servers) > 0 {
		select {
		case <-h.closedNotify:
		case <-time.After(time.Second * 3):
		}
	}
	registry.Close()
	time.Sleep(time.Millisecond * 500)
	if h.watcher != nil {
		h.watcher.Close()
	}
	logger.Close()
	close(h.closedNotify)
}

//freeMemory 每120秒执行1次垃圾回收，清理内存
func (h *Hydra) freeMemory() {
	for {
		select {
		case <-time.After(time.Second * 120):
			debug.FreeOSMemory()
		}
	}
}

//recovery 处理异常恢复
func (h *Hydra) recovery() {
	if e := recover(); e != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "hydra执行期间出现异常: %v", e)
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			} else {
				fmt.Fprintf(&buf, "\n")
			}
			fmt.Fprintf(&buf, "%v:%v", file, line)
		}
		var content = buf.String()
		h.Error(content)
	}
}
