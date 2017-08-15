package web

import (
	"fmt"
	"time"

	"sync"

	"strings"

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
	handler  context.Handler
	mu       sync.Mutex
}

//newHydraAPIServer 创建web服务器
func newHydraWebServer(handler context.Handler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraWebServer, err error) {
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
	routers, err := conf.GetNodeWithSectionName("router")
	if err != nil {
		return fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSectionName("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		baseArgs := routers.String("args")
		rts, err := routers.GetSections("routers")
		if err != nil || len(rts) == 0 {
			return fmt.Errorf("routers路由未配置或配置有误:%s(len:%d,err:%+v)", conf.String("name"), len(rts), err)
		}
		apiRouters := make([]*api.WebRouter, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			actions := strings.Split(strings.ToUpper(c.String("action", "get,post")), ",")
			mode := c.String("mode", "*")
			args := c.String("args")
			if name == "" || service == "" {
				return fmt.Errorf("路由配置错误:service 和 name不能为空（name:%s，service:%s）", name, service)
			}
			for _, v := range actions {
				exist := false
				for _, e := range api.SupportMethods {
					if v == e {
						exist = true
						break
					}
				}
				if !exist {
					return fmt.Errorf("路由配置错误:action:%v不支持,只支持:%v", actions, api.SupportMethods)
				}
			}
			apiRouters = append(apiRouters, &api.WebRouter{
				Method:      actions,
				Path:        name,
				Handler:     w.handle(name, mode, service, baseArgs+"&"+args),
				Middlewares: make([]api.Handler, 0, 0)})
		}
		w.server.SetRouters(apiRouters...)
		//设置通用头信息
		headers, err := routers.GetIMap("headers")
		if err == nil {
			nheader := make(map[string]string)
			for k, v := range headers {
				nheader[k] = fmt.Sprint(v)
			}
			w.server.SetHeader(nheader)
		}
		//设置静态文件路由
		staticConf, err := routers.GetSection("static")
		if err == nil {
			w.server.Infof("%s:启用静态文件", conf.String("name"))
			prefix := staticConf.String("prefix")
			dir := staticConf.String("dir")
			showDir := staticConf.String("showDir") == "true"
			exts := staticConf.Strings("exts")
			if dir == "" {
				return fmt.Errorf("static配置错误：%s,dir,exts不能为空(%s)", conf.String("name"), dir)
			}
			w.server.SetStatic(prefix, dir, showDir, exts)
		}
		//设置xsrf参数，并启用xsrf校验
		xsrf, err := routers.GetSection("xsrf")
		if err == nil {
			w.server.Infof("%s:启用xsrf校验", conf.String("name"))
			key := xsrf.String("key")
			secret := xsrf.String("secret")
			if key == "" || secret == "" {
				return fmt.Errorf("xsrf配置错误：key,secret不能为空(%s,%s,%s)", conf.String("name"), key, secret)
			}
			w.server.SetXSRF(key, secret)
		}
		allowAjax := routers.String("onlyAllowAjaxRequest", "false") == "true"
		if allowAjax {
			w.server.Infof("%s:启用ajax调用限制", conf.String("name"))
		}
		w.server.OnlyAllowAjaxRequest(allowAjax)
	}
	if conf.Has("view") {
		view, err := conf.GetNodeWithSectionName("view")
		if err != nil {
			return fmt.Errorf("view未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSectionName("view"); err != nil || r.GetVersion() != view.GetVersion() {
			path := view.String("viewPath", "../views")
			left := view.String("left", "{{")
			right := view.String("right", "}}")
			w.server.SetViewsPath(path)
			w.server.SetDelims(left, right)
		}
	}
	//设置metric服务器监控数据
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
			w.server.SetInfluxMetric(host, dataBase, userName, password, time.Second*60)
		}
	} else {
		w.server.StopInfluxMetric()
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
		ext["__func_http_request_"] = c.Req()
		ext["__func_http_response_"] = c.ResponseWriter
		ext["__func_body_get_"] = func(ch string) (string, error) {
			return encoding.Convert(c.BodyBuffer, ch)
		}
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#@domain/var/%s/%s", c, n), false)
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
		if response == nil {
			response = context.GetStandardResponse()
		}
		defer func() {
			if err != nil {
				c.Errorf("web.response.error: %v", err)
			}
		}()
		//处理头信息
		for k, v := range response.GetHeaders() {
			c.Header().Set(k, v)
		}
		if err != nil {
			c.Result = err
			return
		}
		if response.IsRedirect() {
			c.Redirect(response.GetParams()["Location"].(string), response.GetStatus(nil))
			return
		}
		c.Result = response
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
	routers, err := conf.GetNodeWithSectionName("router")
	if err != nil {
		return false, fmt.Errorf("路由未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("router"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	if conf.Has("view") {
		view, err := conf.GetNodeWithSectionName("view")
		if err != nil {
			return false, fmt.Errorf("view未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		if r, err := w.conf.GetNodeWithSectionName("view"); err != nil || r.GetVersion() != view.GetVersion() {
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

func (h *webServerAdapter) Resolve(c context.Handler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraWebServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_WEB, &webServerAdapter{})
}
