package hydra

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/qxnw/hydra/server"

	"sync"

	"runtime/debug"

	"github.com/qxnw/hydra/conf"

	"github.com/qxnw/hydra/registry"

	log "github.com/qxnw/hydra/logger"
	"github.com/qxnw/lib4go/logger"
	"github.com/spf13/pflag"
)

//Hydra Hydra server
type Hydra struct {
	system string

	collectIndex int

	*logger.Logger
	watcher      conf.Watcher
	servers      map[string]*Server
	notify       chan *conf.Updater
	closedNotify chan struct{}
	closeChan    chan struct{}
	done         bool
	mu           sync.Mutex
	*HFlags
}

//NewHydra 初始化Hydra服务
func NewHydra() *Hydra {
	h := &Hydra{
		servers:      make(map[string]*Server),
		Logger:       logger.GetSession("hydra", logger.CreateSession()),
		closedNotify: make(chan struct{}, 1),
		closeChan:    make(chan struct{}),
		HFlags:       &HFlags{},
	}
	h.HFlags.BindFlags(pflag.CommandLine)
	return h
}

//Start 启动服务器
func (h *Hydra) Start() (err error) {
	defer h.recovery()
	if err = h.CheckFlags(); err != nil {
		h.Error(err)
		return
	}
	server.IsDebug = h.IsDebug
	if !server.IsDebug {
		logger.AddWriteThread(49) //非调试模式时设置日志写协程数为50个
	}
	//检查是否配置RPC日志服务
	if h.rpcLogger && h.runMode != modeStandalone {
		err = log.ConfigRPCLogger(h.Domain, h.currentRegistry, h.Logger)
		if err != nil {
			h.Errorf("无法启用RPC日志:%v", err)
			return
		}
		h.Info("hydra:启用RPC日志")
	}
	//启动服务器状态查询服务
	if err = h.StartStatusServer(h.Domain); err != nil {
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
	h.watcher, err = conf.NewWatcher(h.runMode, h.Domain, h.tag, h.Logger, h.currentRegistryAddress)
	if err != nil {
		h.Error(fmt.Sprintf("watcher初始化失败 run mode:%s,domain:%s(err:%v)", h.runMode, h.Domain, err))
		return
	}
	h.notify = h.watcher.Notify()
	err = h.watcher.Start()
	if err != nil {
		h.Errorf("watcher启用失败 run mode:%s,domain:%s(err:%v)", h.runMode, h.Domain, err)
		return
	}
	go h.loopRecvNotify()
	go h.freeMemory()
	h.Infof("启动 hydra server(%s,%s)...", h.tag, h.runMode)

	//监听操作系统事件ctrl+c, kill
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM) //9:kill/SIGKILL,15:SIGTEM,20,SIGTOP 2:interrupt/syscall.SIGINT
LOOP:
	for {
		select {
		case <-interrupt:
			h.Warnf("hydra server(%s) was killed", h.Domain)
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
	srv := NewHydraServer(h.Domain, h.runMode, h.currentRegistry, h.currentRegistryAddress, h.crossRegistryAddress, h.Logger)
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
			//case <-time.After(time.Second * 2):
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
