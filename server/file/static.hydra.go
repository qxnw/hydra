package file

import (
	"fmt"
	"time"

	"sync"

	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/net"
)

//hydraWebServer web server适配器
type hydraFileServer struct {
	server   *api.WebServer
	conf     conf.Conf
	registry server.IServiceRegistry
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraWebServer 构建基本配置参数的web server
func newHydraFileServer(handler context.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraFileServer, err error) {
	h = &hydraFileServer{handler: handler,
		registry: r,
		conf:     conf.NewJSONConfWithEmpty(),
		server: api.New(cnf.String("domain"), cnf.String("name", "api.server"),
			api.WithRegistry(r, cnf.Translate("{@category_path}/servers/{@tag}")),
			api.WithIP(net.GetLocalIPAddress(cnf.String("mask"))))}
	err = h.setConf(cnf)
	return
}

func (w *hydraFileServer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.server = api.New(cnf.String("domain"), cnf.String("name", "api.server"),
		api.WithRegistry(w.registry, cnf.Translate("{@category_path}/servers/{@tag}")),
		api.WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraFileServer) setConf(conf conf.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}
	//设置路由
	routers, err := conf.GetNodeWithSection("router")
	if err != nil {
		return fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSection("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		staticConf, err := routers.GetSection("static")
		if err != nil {
			return fmt.Errorf("static未配置:%s(err:%+v)", conf.String("name"), err)
		}
		prefix := staticConf.String("prefix")
		dir := staticConf.String("dir")
		showDir := staticConf.String("showDir") == "true"
		exts := staticConf.Strings("exts")
		if dir == "" {
			return fmt.Errorf("static配置错误：%s,dir,exts不能为空(%s)", conf.String("name"), dir)
		}
		w.server.SetStatic(prefix, dir, showDir, exts)
	}

	//设置metric上报
	if conf.Has("metric") {
		metric, err := conf.GetNodeWithSection("metric")
		if err != nil {
			return fmt.Errorf("metric未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSection("metric"); err != nil || r.GetVersion() != metric.GetVersion() {
			host := metric.String("host")
			dataBase := metric.String("dataBase")
			userName := metric.String("userName")
			password := metric.String("password")
			if host == "" || dataBase == "" {
				return fmt.Errorf("metric配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
			}
			if !strings.Contains(host, "://") {
				host = "http://" + host
			}
			w.server.SetInfluxMetric(host, dataBase, userName, password, time.Second*60)
		}
	} else {
		w.server.StopInfluxMetric()
	}

	//设置基本参数
	w.server.SetHost(conf.String("host"))
	w.conf = conf
	return nil
}

//GetAddress 获取服务器地址
func (w *hydraFileServer) GetAddress() string {
	return w.server.GetAddress()
}
func (w *hydraFileServer) GetStatus() string {
	if w.server.Running {
		return server.ST_RUNNING
	}
	return server.ST_STOP
}

//Start 启用服务
func (w *hydraFileServer) Start() (err error) {
	tls, err := w.conf.GetSection("tls")
	if err != nil {
		go func() {
			err = w.server.Run(w.conf.String("address", ":9898"))
		}()
	} else {
		go func(tls conf.Conf) {
			err = w.server.RunTLS(tls.String("cert"), tls.String("key"), tls.String("address", ":9898"))
		}(tls)
	}
	time.Sleep(time.Second)
	return nil
}

//接口服务变更通知
func (w *hydraFileServer) Notify(conf conf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}

	restart, err := w.needRestart(conf)
	if err != nil {
		return err
	}
	if restart { //服务器地址已变化，则重新启动新的server,并停止当前server
		return w.restartServer(conf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	return w.setConf(conf)
}
func (w *hydraFileServer) needRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	if w.conf.String("address") != conf.String("address") {
		return true, nil
	}
	if w.conf.String("host") != conf.String("host") {
		return true, nil
	}
	routers, err := conf.GetNodeWithSection("router")
	if err != nil {
		return false, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSection("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	return false, nil
}

//Shutdown 关闭服务
func (w *hydraFileServer) Shutdown() {
	timeout, _ := w.conf.Int("timeout", 10)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

type hydraFileServerAdapter struct {
}

func (h *hydraFileServerAdapter) Resolve(c context.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraFileServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_FILE_API, &hydraFileServerAdapter{})
}
