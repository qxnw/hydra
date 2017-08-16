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
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web server适配器
type hydraRPCServer struct {
	server   *RPCServer
	registry server.IServiceRegistry
	conf     conf.Conf
	handler  context.Handler
	mu       sync.Mutex
}

//newHydraRPCServer 构建RPC服务器
func newHydraRPCServer(handler context.Handler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraRPCServer, err error) {
	h = &hydraRPCServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		server: NewRPCServer(cnf.String("domain"), cnf.String("name", "rpc.server"),
			WithRegistry(r, cnf.Translate("{@category_path}/servers")),
			WithIP(net.GetLocalIPAddress(cnf.String("mask")))),
	}
	err = h.setConf(cnf)
	return
}

//restartServer 重启RPC服务器
func (w *hydraRPCServer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.server = NewRPCServer(cnf.String("domain"), cnf.String("name", "rpc.server"),
		WithRegistry(w.registry, cnf.Translate("{@category_path}/servers")),
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
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
	//检查版本是否发生变化
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查服务器状态是否正确
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}
	//设置路由
	routers, err := conf.GetNodeWithSectionName("router")
	if err != nil {
		return fmt.Errorf("router未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSectionName("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		baseArgs := routers.String("args")
		rts, err := routers.GetSections("routers")
		if err != nil {
			return fmt.Errorf("routers未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		routers := make([]*rpcRouter, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			action := strings.Split(strings.ToUpper(c.String("action", "request")), ",")
			args := c.String("args")
			mode := c.String("mode", "*")
			service := c.String("service")
			if name == "" || service == "" {
				return fmt.Errorf("router配置错误:service 和 name不能为空（name:%s，service:%s）", name, service)
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
					return fmt.Errorf("router配置错误:method:%v不支持,只支持:%v", action, SupportMethods)
				}
			}
			routers = append(routers, &rpcRouter{
				Method:      action,
				Path:        name,
				Handler:     w.handle(name, mode, service, baseArgs+"&"+args),
				Middlewares: make([]Handler, 0, 0)})
		}
		w.server.SetRouters(routers...)
	}

	//设置metric服务器监控状态
	if conf.Has("metric") {
		metric, err := conf.GetNodeWithSectionName("metric")
		if err != nil {
			return fmt.Errorf("metric未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSectionName("metric"); err != nil || r.GetVersion() != metric.GetVersion() {
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
			w.server.SetInfluxMetric(host, dataBase, userName, password, 60*time.Second)
		}
	} else {
		w.server.StopInfluxMetric()
	}

	//设置限流规则
	if conf.Has("limiter") {
		limiter, err := conf.GetNodeWithSectionName("limiter")
		if err != nil {
			return fmt.Errorf("limiter未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSectionName("limiter"); err != nil || r.GetVersion() != limiter.GetVersion() {
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
	} else {
		w.server.UpdateLimiter(make(map[string]int))
	}

	//设置基本参数
	w.conf = conf
	return nil
}

//setRouter 设置路由
func (w *hydraRPCServer) handle(name string, mode string, service string, args string) func(c *Context) {
	return func(c *Context) {
		//处理输入参数
		ctx := context.GetContext()
		defer ctx.Close()

		tfParams := transform.New()
		c.Params().Each(func(k, v string) {
			tfParams.Set(k[1:], v)
		})
		tfParams.Set("method", c.Method())
		input := c.Req().GetArgs()
		if raw, ok := input["__raw__"]; ok {
			rawMap, err := jsons.Unmarshal([]byte(raw))
			if err != nil {
				c.Result = &StatusResult{Code: 500, Result: fmt.Errorf("输入参数__raw__必须是json：%v", err), Type: JsonResponse}
				return
			}
			smap, err := types.ToStringMap(rawMap)
			if err != nil {
				c.Result = &StatusResult{Code: 500, Result: fmt.Errorf("输入参数__raw__有误：%v", err), Type: JsonResponse}
				return
			}
			for k, v := range smap {
				input[k] = v
			}
		}
		tfForm := transform.NewMap(input)

		rArgs := tfForm.Translate(tfParams.Translate(args))
		body, _ := tfForm.Get("__body")
		ext := map[string]interface{}{"hydra_sid": c.GetSessionID()}

		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#@domain/var/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return string(cnf), nil
		}
		margs, err := utility.GetMapWithQuery(rArgs)
		if err != nil {
			c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf("err:%+v", err.Error()), Type: AutoResponse}
			return
		}

		ctx.SetInput(tfForm.Data, tfParams.Data, body, margs, ext)

		//执行服务调用
		response, err := w.handler.Handle(name, mode, c.Req().Service, ctx)
		if response == nil {
			response = context.GetStandardResponse()
		}
		defer func() {
			if err != nil {
				c.Errorf("rpc.response.error: %v", err)
			}
		}()

		//处理错误err,5xx
		if err != nil {
			err = fmt.Errorf("rpc.server.handler.error:%v", err)
			if server.IsDebug {
				c.Result = &StatusResult{Code: response.GetStatus(err), Result: response.GetContent(err), Type: response.GetContentType()}
				return
			}
			err = errors.New("Internal Server Error(工作引擎发生异常)")
			c.Result = &StatusResult{Code: response.GetStatus(err), Result: err, Type: response.GetContentType()}
			return
		}
		c.Result = &StatusResult{Code: response.GetStatus(), Result: response.GetContent(), Params: response.GetParams(), Type: response.GetContentType()}
	}
}

//GetAddress 获取服务器地址
func (w *hydraRPCServer) GetAddress() string {
	return w.server.GetAddress()
}

//Start 启用服务
func (w *hydraRPCServer) Start() (err error) {
	return w.server.Start(w.conf.String("address", ":9899"))
}

//接口服务变更通知
func (w *hydraRPCServer) Notify(conf conf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查任务列表等是否变化，判断是否需要重启
	restart, err := w.needRestart(conf)
	if err != nil {
		return err
	}
	if restart {
		return w.restartServer(conf)
	}
	//任务列表无变化
	return w.setConf(conf)
}
func (w *hydraRPCServer) needRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	if w.conf.String("address") != conf.String("address") {
		return true, nil
	}
	routers, err := conf.GetNodeWithSectionName("router")
	if err != nil {
		return false, fmt.Errorf("queue未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	return false, nil
}
func (w *hydraRPCServer) GetStatus() string {
	if w.server.running {
		return server.ST_RUNNING
	}
	return server.ST_STOP
}

//Shutdown 关闭服务
func (w *hydraRPCServer) Shutdown() {
	w.server.Close()
}

type hydraRPCServerAdapter struct {
}

func (h *hydraRPCServerAdapter) Resolve(c context.Handler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraRPCServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_RPC, &hydraRPCServerAdapter{})
}
