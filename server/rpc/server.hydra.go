package rpc

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
)

//hydraWebServer web server适配器
type hydraRPCServer struct {
	server   *RPCServer
	registry context.IServiceRegistry
	conf     registry.Conf
	logger   context.Logger
	handler  context.EngineHandler
	versions map[string]int32
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraRPCServer(handler context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (h *hydraRPCServer, err error) {
	h = &hydraRPCServer{handler: handler,
		logger:   logger,
		registry: r,
		versions: make(map[string]int32),
		server: NewRPCServer(conf.String("name", "rpc.server"),
			WithRegistry(r),
			WithLogger(logger),
			WithIP(net.GetLocalIPAddress(conf.String("mask")))),
	}
	err = h.setConf(conf)
	return
}

//restartServer 重启服务器
func (w *hydraRPCServer) restartServer(conf registry.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	for k := range w.versions {
		delete(w.versions, k)
	}
	w.server = NewRPCServer(conf.String("name", "rpc.server"),
		WithRegistry(w.registry),
		WithLogger(w.logger),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	err = w.setConf(conf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraRPCServer) setConf(conf registry.Conf) error {
	if w.conf == nil {
		w.conf = registry.NewJSONConfWithEmpty()
	}
	if  w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置路由
	routers, err := conf.GetNode("router")
	if err != nil {
		return fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, ok := v.conf.GetNode("router"); ok && r.eGtVersion() != routers.GetVersion()|| !ok  {
		w.versions["routers"] = routers.GetVersion()
		rts, err := routers.GetSections("routers")
		if err != nil {
			return err
		}
		routers := make([]*rpcRouter, 0, len(rts))
		for _, c := range rts {
			method := strings.Split(strings.ToUpper(c.String("method", "request")), ",")
			service := c.String("service")
			params := c.String("params")
			name := c.String("name")
			if name == "" || service == "" {
				return fmt.Errorf("路由配置错误:service 和 name不能为空（name:%s，service:%s）", name, service)
			}
			for _, v := range method {
				exist := false
				for _, e := range SupportMethods {
					if v == e {
						exist = true
						break
					}
				}
				if !exist {
					return fmt.Errorf("路由配置错误:method:%v不支持,只支持:%v", method, SupportMethods)
				}
			}

			routers = append(routers, &rpcRouter{
				Method:      method,
				Path:        name,
				Handler:     w.handle(name, method, service, params),
				Middlewares: make([]Handler, 0, 0)})
		}
		w.server.SetRouters(routers...)
	}

	//设置metric上报
	metric, err := conf.GetNode("metric")
	if r, ok := v.conf.GetNode("metric"); ok && r.eGtVersion() != metric.GetVersion() || !ok {
		host := metric.String("host")
		dataBase := metric.String("dataBase")
		userName := metric.String("userName")
		password := metric.String("password")
		timeSpan, _ := metric.Int("timeSpan", 10)
		if host == "" || dataBase == "" {
			return fmt.Errorf("metric配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		w.server.SetInfluxMetric(host, dataBase, userName, password, time.Duration(timeSpan)*time.Second)
	}
	limiter, err := conf.GetNode("limiter")
	if r, ok := v.conf.GetNode("limiter"); ok && r.eGtVersion() != limiter.GetVersion() || !ok {
		w.versions["limiter"] = limiter.GetVersion()
		lmts, err := limiter.GetSections("qos")
		if err != nil {
			return err
		}
		limiters:=map[string]int
		for _,v:=range lmts{
			name:=v.String("name")
			lm,err:=v.Int("value")
			if err!=nil{
				return fmt.Errorf("limiter配置错误:qos.value值必须为数字（%s）", v.String(max))
			}
			limiters[name]=lm
		}
		w.server.UpdateLimiter(limiters)
	}

	//设置基本参数
	w.server.SetName(conf.String("name", "rpc.server"))
	w.conf = conf
	return nil
}

//setRouter 设置路由
func (w *hydraRPCServer) handle(name string, method []string, service string, params string) func(c *Context) {
	return func(c *Context) {

		//处理输入参数
		hydraContext := make(map[string]interface{})
		tfParams := transform.NewGetter(c.Params())
		tfForm := transform.NewMap(c.Req().GetArgs())

		rparams := tfForm.Translate(tfParams.Translate(params))
		hydraContext["__func_param_getter_"] = tfParams
		hydraContext["__func_args_getter_"] = tfForm

		//执行服务调用
		response, err := w.handler.Handle(name, c.Method(), c.Req().Service, rparams, hydraContext)
		if err != nil {
			c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf("err:%+v", err.Error()), Type: 0}
			return
		}

		//处理返回参数
		if response.Status == 0 {
			response.Status = 200
		}
		c.Result = &StatusResult{Code: response.Status, Result: response.Content, Type: JsonResponse}
	}
}

//GetAddress 获取服务器地址
func (w *hydraRPCServer) GetAddress() string {
	return w.server.GetAddress()
}

//Start 启用服务
func (w *hydraRPCServer) Start() (err error) {
	err = w.server.Start(w.conf.String("address", ":9899"))
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	return nil
}

//接口服务变更通知
func (w *hydraRPCServer) Notify(conf registry.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil && w.conf.GetVersion() == conf.GetVersion() {
		return errors.New("版本无变化")
	}
	if w.conf != nil && w.conf.String("address") != conf.String("address") { //服务器地址已变化，则重新启动新的server,并停止当前server
		return w.restartServer(conf)
	}
	//服务器地址未变化，更新服务器当前配置，并立即生效
	return w.setConf(conf)
}

//Shutdown 关闭服务
func (w *hydraRPCServer) Shutdown() {
	w.server.Close()
}

type hydraRPCServerAdapter struct {
}

func (h *hydraRPCServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (server.IHydraServer, error) {
	return newHydraRPCServer(c, r, conf, logger)
}

func init() {
	server.Register("rpc.server", &hydraRPCServerAdapter{})
}
