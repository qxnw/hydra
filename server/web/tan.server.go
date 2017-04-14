package web

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/encoding"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web server适配器
type hydraWebServer struct {
	server   *WebServer
	conf     conf.Conf
	registry context.IServiceRegistry
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraWebServer 构建基本配置参数的web server
func newHydraWebServer(handler context.EngineHandler, r context.IServiceRegistry, cnf conf.Conf) (h *hydraWebServer, err error) {
	h = &hydraWebServer{handler: handler,
		registry: r,
		conf:     conf.NewJSONConfWithEmpty(),
		server: New(cnf.String("name", "api.server"),
			WithRegistry(r),
			WithIP(net.GetLocalIPAddress(cnf.String("mask"))))}
	err = h.setConf(cnf)
	return
}

func (w *hydraWebServer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.server = New(cnf.String("name", "api.server"),
		WithRegistry(w.registry),
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraWebServer) setConf(conf conf.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置路由
	routers, err := conf.GetNode("router")
	if err != nil {
		return fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNode("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("routers")
		if err != nil || len(rts) == 0 {
			return fmt.Errorf("routers路由未配置或配置有误:%s(len:%d,err:%+v)", conf.String("name"), len(rts), err)
		}
		routers := make([]*webRouter, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			method := strings.Split(strings.ToUpper(c.String("method", "post")), ",")
			params := c.String("params")
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

			routers = append(routers, &webRouter{
				Method:      method,
				Path:        name,
				Handler:     w.handle(name, method, service, params),
				Middlewares: make([]Handler, 0, 0)})
		}
		w.server.SetRouters(routers...)
	}

	//设置metric上报
	metric, err := conf.GetNode("metric")
	if err != nil {
		return fmt.Errorf("metric未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNode("metric"); err != nil || r.GetVersion() != metric.GetVersion() {
		host := metric.String("host")
		dataBase := metric.String("dataBase")
		userName := metric.String("userName")
		password := metric.String("password")
		//timeSpan, _ := metric.Int("timeSpan", 5)
		if host == "" || dataBase == "" {
			return fmt.Errorf("metric配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		w.server.SetInfluxMetric(host, dataBase, userName, password, time.Second)
	}
	//设置基本参数
	w.server.SetName(conf.String("name", "api.server"))
	w.server.SetHost(conf.String("host"))
	w.conf = conf
	return nil
}

//setRouter 设置路由
func (w *hydraWebServer) handle(name string, method []string, service string, args string) func(c *Context) {
	return func(c *Context) {

		//处理输入参数
		context := context.GetContext()
		defer context.Close()
		buf, err := c.Body()
		if err != nil {
			c.BadRequest(fmt.Sprintf("%+v", err))
			return
		}
		tfParams := transform.NewGetter(c.Params())
		tfForm := transform.NewGetter(c.Forms().Form)
		context.Ext["hydra_sid"] = c.GetSessionID()
		context.Ext["__txt_body_"] = string(buf)
		context.Ext["__func_param_getter_"] = tfParams
		context.Ext["__func_args_getter_"] = tfForm
		context.Ext["__func_http_request_"] = c.Req()
		context.Ext["__func_http_response_"] = c.ResponseWriter
		context.Ext["__func_body_get_"] = func(c string) (string, error) {
			return encoding.Convert(buf, c)
		}
		rservice := tfForm.Translate(tfParams.Translate(service))
		rArgs := tfForm.Translate(tfParams.Translate(args))

		context.Input.Input = tfForm.Data
		context.Input.Body = string(buf)
		context.Input.Params = tfParams.Data

		context.Input.Args, err = utility.GetMapWithQuery(rArgs)
		if err != nil {
			c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf("err:%+v", err.Error()), Type: 0}
			return
		}
		//执行服务调用
		response, err := w.handler.Handle(name, c.Req().Method, rservice, context)
		if err != nil {
			c.Result = &StatusResult{Code: 500, Result: fmt.Sprintf("err:%+v", err.Error()), Type: 0}
			return
		}

		//处理返回参数
		for k, v := range response.Params {
			c.Header().Set(k, v.(string))
		}
		if response.Status == 0 {
			response.Status = 200
		}
		var typeID = JsonResponse
		if tp, ok := response.Params["Content-Type"].(string); ok {
			if strings.Contains(tp, "xml") {
				typeID = XmlResponse
			} else if strings.Contains(tp, "plain") {
				typeID = AutoResponse
			}
		}
		c.Result = &StatusResult{Code: response.Status, Result: response.Content, Type: typeID}
	}
}

//GetAddress 获取服务器地址
func (w *hydraWebServer) GetAddress() string {
	return w.server.GetAddress()
}

//Start 启用服务
func (w *hydraWebServer) Start() (err error) {
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
func (w *hydraWebServer) Notify(conf conf.Conf) error {
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
func (w *hydraWebServer) Shutdown() {
	timeout, _ := w.conf.Int("timeout", 5)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

type hydraWebServerAdapter struct {
}

func (h *hydraWebServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraWebServer(c, r, conf)
}

func init() {
	server.Register("api", &hydraWebServerAdapter{})
}
