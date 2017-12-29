package web

import (
	"fmt"
	"time"

	"sync"

	"strings"

	"reflect"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/hydra/server/api"
	"github.com/qxnw/lib4go/encoding"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web 服务器
type hydraWebServer struct {
	server   *WebServer
	conf     conf.Conf
	registry server.IServiceRegistry
	handler  server.EngineHandler
	mu       sync.Mutex
}

//newHydraAPIServer 创建web服务器
func newHydraWebServer(handler server.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraWebServer, err error) {
	h = &hydraWebServer{handler: handler,
		registry: r,
		conf:     conf.NewJSONConfWithEmpty(),
		server: New(cnf.String("domain"), cnf.String("name", "web.server"),
			api.WithRegistry(r, cnf.Translate("{@category_path}/servers/{@tag}")),
			api.WithIP(net.GetLocalIPAddress(cnf.String("mask"))))}
	err = h.setConf(cnf)
	return
}

//restartServer 重启服务器
func (w *hydraWebServer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	time.Sleep(time.Second)
	w.server = New(cnf.String("domain"), cnf.String("name", "web.server"),
		api.WithRegistry(w.registry, cnf.Translate("{@category_path}/servers/{@tag}")),
		api.WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//setConf 设置配置参数
func (w *hydraWebServer) setConf(conf conf.Conf) error {
	//检查版本号
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查服务器状态
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}

	//设置路由
	routers, err := server.GetRouters(w.conf, conf, "get", api.SupportMethods)
	if err != nil && err != server.ERR_NO_CHANGED {
		err = fmt.Errorf("路由配置有误:%v", err)
		return err
	}
	if err == nil {
		apiRouters := make([]*api.WebRouter, 0, len(routers))
		for _, router := range routers {
			apiRouters = append(apiRouters, &api.WebRouter{
				Method:      router.Action,
				Path:        router.Name,
				Handler:     w.handle(router.Name, router.Mode, router.Service, router.Args),
				Middlewares: make([]api.Handler, 0, 0)})
		}
		w.server.SetRouters(apiRouters...)
	}

	//设置静态文件路由
	enable, prefix, dir, showDir, exts, err1 := server.GetStatic(w.conf, conf)
	if err1 != nil && err1 != server.ERR_NO_CHANGED && err1 != server.ERR_NOT_SETTING {
		return err1
	}
	if err1 == server.ERR_NOT_SETTING || !enable {
		w.server.Infof("%s(%s):未配置静态文件", conf.String("name"), conf.String("type"))
		w.server.SetStatic(false, prefix, dir, showDir, exts)
	}
	if err1 == nil && enable {
		w.server.Infof("%s(%s):启用静态文件", conf.String("name"), conf.String("type"))
		w.server.SetStatic(true, prefix, dir, showDir, exts)
	}
	if err1 != nil && err != nil {
		return fmt.Errorf("路由配置有误:%v，静态文件:%v", err, err1)
	}
	if len(routers) == 0 {
		w.server.Infof("%s(%s):未配置路由", conf.String("name"), conf.String("type"))
	}

	//设置通用头信息
	headers, err := server.GetHeaders(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == nil || err == server.ERR_NOT_SETTING {
		w.server.Infof("%s(%s):设置头%d", conf.String("name"), conf.String("type"), len(headers))
		w.server.SetHeader(headers)
	}

	//设置view配置
	views, err := server.GetViews(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		return err
	}
	if err == server.ERR_NOT_SETTING {
		w.server.SetViewsPath("../views")
		w.server.SetDelims("{{", "}}")
	}
	if err == nil {
		w.server.Infof("%s(%s):启用自定义view路径%s", conf.String("name"), conf.String("type"), views.ViewPath)
		w.server.SetViewsPath(views.ViewPath)
		w.server.SetDelims(views.Left, views.Right)
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

	//设置basic安全认证参数
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

	//设置其它参数
	w.server.SetHost(conf.String("host"))
	w.conf = conf
	return nil
}

//handle api请求处理程序
func (w *hydraWebServer) handle(name string, mode string, service string, args string) func(c *api.Context) {
	return func(c *api.Context) {
		//处理输入参数
		ctx := context.GetContext()
		defer ctx.Close()

		ext := make(map[string]interface{})
		ext["hydra_sid"] = c.GetSessionID()
		ext["__jwt_"] = c.GetJWTStorage()
		ext["__func_http_request_"] = c.Req()
		ext["__func_http_response_"] = c.ResponseWriter
		ext["__func_body_get_"] = func(ch string) (string, error) {
			return encoding.Convert(c.BodyBuffer, ch)
		}
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return string(cnf), nil
		}

		tfParams := transform.New()
		c.Params().Each(func(k, v string) {
			tfParams.Set(k[1:], v)
		})
		tfParams.Set("method", c.Req().Method)
		tfForm := transform.NewValues(c.Forms().Form)
		c.ServiceName = tfForm.Translate(tfParams.Translate(service))
		rArgs := tfForm.Translate(tfParams.Translate(args))
		margs, err := utility.GetMapWithQuery(rArgs)
		if err != nil {
			return
		}

		ctx.SetInput(tfForm.Data, tfParams.Data, string(c.BodyBuffer), margs, ext)

		//调用执行引擎进行逻辑处理
		response, err := w.handler.Handle(name, mode, c.ServiceName, ctx)
		if reflect.ValueOf(response).IsNil() {
			response = context.GetStandardResponse()
		}
		defer func() {
			if err != nil {
				c.Errorf("web.response.error: %v", err)
			}
		}()
		if err != nil {
			response.SetError(response.GetStatus(err), err)
		}
		//处理头信息
		for k, v := range response.GetHeaders() {
			c.Header().Set(k, v)
		}

		//设置jwt.token
		c.SetJwtToken(response.GetParams()["__jwt_"])
		c.Result = &webResult{Response: response, Error: err, Mode: mode}
		if err != nil {
			return
		}

		if url, ok := response.IsRedirect(); ok {
			c.Redirect(url, response.GetStatus())
			return
		}

	}
}

//GetAddress 获取服务器地址
func (w *hydraWebServer) GetAddress() string {
	return w.server.GetAddress()
}

//GetStatus 获取当前服务器状态
func (w *hydraWebServer) GetStatus() string {
	if w.server.Running {
		return server.ST_RUNNING
	}
	return server.ST_STOP
}
func (w *hydraWebServer) GetServices() []string {
	return w.handler.GetServices()
}

//Start 启用服务
func (w *hydraWebServer) Start() (err error) {
	tls, err := w.conf.GetSection("tls")
	startChan := make(chan error, 1)
	if err != nil {
		go func(ch chan error) {
			err = w.server.Run(w.conf.String("address", ":9898"))
			startChan <- err
		}(startChan)
	} else {
		go func(tls conf.Conf, ch chan error) {
			err = w.server.RunTLS(tls.String("cert"), tls.String("key"), tls.String("address", ":9898"))
			startChan <- err
		}(tls, startChan)
	}
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-startChan:
		return err
	}
}

//Notify 服务器配置变更通知
func (w *hydraWebServer) Notify(conf conf.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查是否需要重启服务器
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

//needRestart 检查配置判断是否需要重启服务器
func (w *hydraWebServer) needRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	if w.conf.String("address") != conf.String("address") {
		return true, nil
	}
	if w.conf.String("host") != conf.String("host") {
		return true, nil
	}
	routers, err := conf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return false, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("router", "#@path/router"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	if conf.Has("#@path/view") {
		view, err := conf.GetNodeWithSectionName("view", "#@path/view")
		if err != nil {
			return false, fmt.Errorf("view未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSectionName("view", "#@path/view"); err != nil || r.GetVersion() != view.GetVersion() {
			return true, nil
		}
	}
	return false, nil
}

//Shutdown 关闭服务器
func (w *hydraWebServer) Shutdown() {
	timeout, _ := w.conf.Int("timeout", 10)
	w.server.Shutdown(time.Duration(timeout) * time.Second)
}

type webServerAdapter struct {
}

func (h *webServerAdapter) Resolve(c server.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraWebServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_WEB, &webServerAdapter{})
}
