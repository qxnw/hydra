package micro

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/qxnw/hydra/servers"

	registry "github.com/qxnw/hydra/registry.v2"
	"github.com/qxnw/hydra/registry.v2/watcher"
	"github.com/qxnw/lib4go/logger"
)

//app  hydra app
type hydra struct {
	logger       *logger.Logger
	closeChan    chan struct{}
	interrupt    chan os.Signal
	isDebug      bool
	platName     string
	systemName   string
	serverTypes  []string
	clusterName  string
	registryAddr string

	mu       sync.Mutex
	registry registry.IRegistry
	watcher  *watcher.ConfWatcher
	notify   chan *watcher.ContentChangeArgs
	servers  *responsiveServers
	trace    string
	done     bool
}

func newHydra(platName string, systemName string, serverTypes []string, clusterName string, trace string, registryAddr string, isDebug bool) *hydra {
	servers.IsDebug = isDebug
	return &hydra{
		logger:       logger.New("hydra"),
		closeChan:    make(chan struct{}),
		interrupt:    make(chan os.Signal, 1),
		isDebug:      isDebug,
		platName:     platName,
		systemName:   systemName,
		serverTypes:  serverTypes,
		clusterName:  clusterName,
		registryAddr: registryAddr,
		trace:        trace,
	}
}

//Start 启动hydra服务器
func (h *hydra) Start() (err error) {

	//非调试模式时设置日志写协程数为50个
	if !h.isDebug {
		logger.AddWriteThread(49)
	}

	//创建注册中心
	if h.registry, err = registry.NewRegistryWithAddress(h.registryAddr, h.logger); err != nil {
		return
	}

	if err = startTrace(h.trace); err != nil {
		return
	}

	if err = h.startWatch(); err != nil {
		return err
	}

	//堵塞当前进程，等服务结束
	signal.Notify(h.interrupt, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1) //9:kill/SIGKILL,15:SIGTEM,20,SIGTOP 2:interrupt/syscall.SIGINT
LOOP:
	for {
		select {
		case <-h.interrupt:
			h.logger.Warnf("hydra 正在安全退出")
			h.done = true
			break LOOP
		}
	}
	h.servers.Shutdown()
	h.logger.Warnf("hydra 已安全退出")
	return nil
}

//startWatch 启动服务器配置监控
func (h *hydra) startWatch() (err error) {
	h.watcher, err = watcher.NewConfWatcher(h.platName, h.systemName, h.serverTypes, h.clusterName, h.registry, h.logger)
	if err != nil {
		err = fmt.Errorf("watcher初始化失败 %s,%+v", filepath.Join(h.platName, h.systemName), err)
		return
	}
	h.logger.Infof("启动 hydra server(%s)...", filepath.Join(h.platName, h.systemName))
	if h.notify, err = h.watcher.Notify(); err != nil {
		return err
	}
	if err != nil {
		err = fmt.Errorf("watcher启动失败 %s,%+v", filepath.Join(h.platName, h.systemName), err)
		return
	}

	//创建服务管理器
	h.servers = newResponsiveServers(h.registry, h.logger)

	go h.loopRecvNotify()
	go h.freeMemory()
	return nil
}

//freeMemory 每120秒执行1次垃圾回收，清理内存
func (h *hydra) freeMemory() {
	for {
		select {
		case <-h.closeChan:
			return
		case <-time.After(time.Second * 120):
			debug.FreeOSMemory()
		}
	}
}

//循环接收服务器配置变化通知
func (h *hydra) loopRecvNotify() {
LOOP:
	for {
		select {
		case <-h.closeChan:
			break LOOP
		case u := <-h.notify:
			if h.done {
				break LOOP
			}
			h.servers.Change(u)
		}
	}
}
func (h *hydra) Shutdown() {
	h.done = true
	close(h.closeChan)
	h.interrupt <- syscall.SIGUSR1

}
