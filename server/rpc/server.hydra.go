package rpc

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web server适配器
type hydraRPCServer struct {
	server   *RPCServer
	registry context.IServiceRegistry
	conf     conf.Conf
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraRPCServer(handler context.EngineHandler, r context.IServiceRegistry, cnf conf.Conf) (h *hydraRPCServer, err error) {
	h = &hydraRPCServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		server: NewRPCServer(cnf.String("name", "rpc.server"),
			WithRegistry(r),
			WithIP(net.GetLocalIPAddress(cnf.String("mask")))),
	}
	err = h.setConf(cnf)
	return
}

//restartServer 重启服务器
func (w *hydraRPCServer) restartServer(conf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.server = NewRPCServer(conf.String("name", "rpc.server"),
		WithRegistry(w.registry),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	err = w.setConf(conf)
	if err != nil {
		return
	}
	err = w.Start()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	return
}

//SetConf 设置配置参数
func (w *hydraRPCServer) setConf(conf conf.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置路由
	routers, err := conf.GetNodeWithSection("router")
	if err != nil {
		return fmt.Errorf("router未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSection("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("routers")
		if err != nil {
			return fmt.Errorf("routers未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		routers := make([]*rpcRouter, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			action := strings.Split(strings.ToUpper(c.String("action", "request")), ",")
			args := c.String("args")
			mode := c.String("mode", "*")
			if name == "" || service == "" {
				return fmt.Errorf("路由配置错误:service 和 name不能为空（name:%s，service:%s）", name, service)
			}
			for _, v := range action {
				exist := false
				for _, e := range SupportMethods {
					if v == e {
						exist = true
						break
					}
				}
				if !exist {
					return fmt.Errorf("路由配置错误:method:%v不支持,只支持:%v", action, SupportMethods)
				}
			}
			routers = append(routers, &rpcRouter{
				Method:      action,
				Path:        name,
				Handler:     w.handle(name, mode, service, args),
				Middlewares: make([]Handler, 0, 0)})
		}
		w.server.SetRouters(routers...)
	}

	//设置metric上报
	metric, err := conf.GetNodeWithSection("metric")
	if err != nil {
		return fmt.Errorf("metric未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSection("metric"); err != nil || r.GetVersion() != metric.GetVersion() {
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
	limiter, err := conf.GetNodeWithSection("limiter")
	if err != nil {
		return fmt.Errorf("limiter未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSection("limiter"); err != nil || r.GetVersion() != limiter.GetVersion() {
		lmts, err := limiter.GetSections("QPS")
		if err != nil {
			return fmt.Errorf("QPS未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		limiters := map[string]int{}
		for _, v := range lmts {
			name := v.String("name")
			lm, err := v.Int("value")
			if err != nil {
				return fmt.Errorf("limiter配置错误:[%s]qos.value:[%s]值必须为数字（err:%v）", name, v.String("value"), err)
			}
			limiters[name] = lm
		}
		w.server.UpdateLimiter(limiters)
	}

	//设置基本参数
	w.server.SetName(conf.String("name", "rpc.server"))
	w.conf = conf
	return nil
}

//setRouter 设置路由
func (w *hydraRPCServer) handle(name string, mode string, service string, args string) func(c *Context) {
	return func(c *Context) {

		//处理输入参数
		context := context.GetContext()
		defer context.Close()
		tfParams := transform.NewGetter(c.Params())
		tfParams.Set("method", c.Method())

		tfForm := transform.NewMap(c.Req().GetArgs())

		rArgs := tfForm.Translate(tfParams.Translate(args))
		context.Ext["hydra_sid"] = c.GetSessionID()
		var err error
		context.Input.Input = tfForm.Data
		context.Input.Params = tfParams.Data
		context.Input.Args, err = utility.GetMapWithQuery(rArgs)
		if err != nil {
			c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf("err:%+v", err.Error()), Type: 0}
			return
		}
		//执行服务调用
		response, err := w.handler.Handle(name, mode, c.Req().Service, context)
		if err != nil {
			if server.IsDebug {
				c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf(":%+v", err.Error()), Type: 0}
				return
			}
			c.Result = &StatusResult{Code: 500, Result: "500 Internal Server Error(工作引擎发生异常)", Type: 0}
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
func (w *hydraRPCServer) Notify(conf conf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf.GetVersion() == conf.GetVersion() {
		return errors.New("版本无变化")
	}
	if w.conf.String("address") != conf.String("address") { //服务器地址已变化，则重新启动新的server,并停止当前server

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

func (h *hydraRPCServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraRPCServer(c, r, conf)
}

func init() {
	server.Register("rpc", &hydraRPCServerAdapter{})
}
