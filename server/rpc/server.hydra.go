package rpc

import (
	"errors"
	"fmt"
	"reflect"
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
	handler  server.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建RPC服务器
func newHydraRPCServer(handler server.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraRPCServer, err error) {
	h = &hydraRPCServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		server: NewRPCServer(cnf.String("domain"), cnf.String("name", "rpc.server"),
			handler.GetServices(),
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
		w.handler.GetServices(),
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
	routers, err := server.GetRouters(w.conf, conf, "request", SupportMethods)
	if err != nil && err != server.ERR_NO_CHANGED {
		err = fmt.Errorf("路由配置有误:%v", err)
		return err
	}
	if err == nil {
		apiRouters := make([]*rpcRouter, 0, len(routers))
		for _, router := range routers {
			apiRouters = append(apiRouters, &rpcRouter{
				Method:      router.Action,
				Path:        router.Name,
				service:     router.Service,
				Handler:     w.handle(router.Name, router.Mode, router.Service, router.Args),
				Middlewares: make([]Handler, 0, 0)})
		}
		w.server.SetRouters(apiRouters...)
	}

	//设置限流规则
	limiters, err := server.GetLimiters(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		err = fmt.Errorf("limiter配置有误:%v", err)
		return err
	}
	if err == nil {
		limitMap := map[string]int{}
		for _, limiter := range limiters {
			limitMap[limiter.Name] = limiter.Value
		}
		if len(limitMap) > 0 {
			w.server.Infof("%s(%s):启用limiter", conf.String("name"), conf.String("type"))
		}
		w.server.UpdateLimiter(limitMap)
	}

	//设置xsrf安全认证参数
	xsrf, err := server.GetAuth(w.conf, conf, "xsrf")
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == server.ERR_NOT_SETTING || !xsrf.Enable {
		w.server.SetXSRF(xsrf.Enable, xsrf.Name, xsrf.Secret, xsrf.Exclude, xsrf.ExpireAt)
	}
	if err == nil && xsrf.Enable {
		w.server.Infof("%s(%s):启用xsrf校验", conf.String("name"), conf.String("type"))
		w.server.SetXSRF(xsrf.Enable, xsrf.Name, xsrf.Secret, xsrf.Exclude, xsrf.ExpireAt)
	}

	//设置jwt安全认证参数
	jwt, err := server.GetAuth(w.conf, conf, "jwt")
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == server.ERR_NOT_SETTING || !jwt.Enable {
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
	}
	if err == nil && jwt.Enable {
		w.server.Infof("%s(%s):启用jwt校验", conf.String("name"), conf.String("type"))
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
	}

	//设置basic安全认证参数,公共secret进行签名
	basic, err := server.GetAuth(w.conf, conf, "basic")
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == server.ERR_NOT_SETTING || !basic.Enable {
		w.server.SetBasic(basic.Enable, basic.Name, basic.Mode, basic.Secret, basic.Exclude, basic.ExpireAt)
	}
	if err == nil && basic.Enable {
		w.server.Infof("%s(%s):启用basic校验", conf.String("name"), conf.String("type"))
		w.server.SetBasic(basic.Enable, basic.Name, basic.Mode, basic.Secret, basic.Exclude, basic.ExpireAt)
	}

	//设置api安全认证参数，独立secret进行签名
	api, err := server.GetAuth(w.conf, conf, "api")
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == server.ERR_NOT_SETTING || !api.Enable {
		w.server.SetAPI(api.Enable, api.Name, api.Mode, api.Secret, api.Exclude, api.ExpireAt)
	}
	if err == nil && api.Enable {
		w.server.Infof("%s(%s):启用api校验", conf.String("name"), conf.String("type"))
		w.server.SetAPI(api.Enable, api.Name, api.Mode, api.Secret, api.Exclude, api.ExpireAt)
	}

	//设置metric服务器监控数据
	enable, host, dataBase, userName, password, span, err := server.GetMetric(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		w.server.Errorf("%s(%s):metric配置有误(%v)", conf.String("name"), conf.String("type"), err)
		w.server.StopInfluxMetric()
	}
	if err == server.ERR_NOT_SETTING || !enable {
		w.server.Warnf("%s(%s):未配置metric", conf.String("name"), conf.String("type"))
		w.server.StopInfluxMetric()
	}
	if err == nil && enable {
		w.server.Infof("%s(%s):启用metric", conf.String("name"), conf.String("type"))
		w.server.SetInfluxMetric(host, dataBase, userName, password, span)
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
		ext["__jwt_"] = c.jwtStorage
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#/@domain/var/%s/%s", c, n), false)
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
		response, err := w.handler.Execute(name, mode, c.Req().Service, ctx)
		if reflect.ValueOf(response).IsNil() {
			response = context.GetStandardResponse()
		}
		defer func() {
			response.Close()
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
	routers, err := conf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return false, fmt.Errorf("router未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("router", "#@path/router"); err != nil || r.GetVersion() != routers.GetVersion() {
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
func (w *hydraRPCServer) GetServices() []string {
	return w.server.remoteRPCService
}

//Shutdown 关闭服务
func (w *hydraRPCServer) Shutdown() {
	w.server.Close()
}

type hydraRPCServerAdapter struct {
}

func (h *hydraRPCServerAdapter) Resolve(c server.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraRPCServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_RPC, &hydraRPCServerAdapter{})
}
