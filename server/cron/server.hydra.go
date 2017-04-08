package cron

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
)

//hydraWebServer web server适配器
type hydraCronServer struct {
	server  *CronServer
	logger  context.Logger
	conf    registry.Conf
	handler context.EngineHandler
	mu      sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraCronServer(handler context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (h *hydraCronServer, err error) {
	h = &hydraCronServer{handler: handler,
		logger: logger,
		server: NewCronServer(conf.String("name", "cron.server"),
			60,
			time.Second,
			WithRegistry(r),
			WithLogger(logger),
			WithIP(net.GetLocalIPAddress(conf.String("mask")))),
	}
	err = h.setConf(conf)
	return
}

//restartServer 重启服务器
func (w *hydraCronServer) restartServer(conf registry.Conf) (err error) {
	w.Shutdown()
	for k := range w.versions {
		delete(w.versions, k)
	}
	w.server = NewCronServer(conf.String("name", "cron.server"),
		60,
		time.Second,
		WithRegister(w.registry),
		WithLogger(w.logger),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	err = w.setConf(conf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraCronServer) setConf(conf registry.Conf) error {
	if w.conf == nil {
		w.conf = registry.NewJSONConfWithEmpty()
	}
	if w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置路由
	routers, err := conf.GetNode("task")
	if err != nil {
		return fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, ok := v.conf.GetNode("task"); ok && r.GetVersion() != routers.GetVersion() || !ok {
		rts, err := routers.GetSections("tasks")
		if err != nil {
			return err
		}
		tasks := make([]*Task, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			params := c.String("params")
			service := c.String("service")
			method := c.String("method")
			interval, err := time.ParseDuration(c.String("interval"))
			if err != nil {
				return fmt.Errorf("task配置错误:interval值必须为整数（%s,%s）(%v)", name, c.String("interval"), err)
			}
			next, err := time.Parse("2006/01/02 15:04:05", c.String("next"))
			if err != nil {
				return fmt.Errorf("task配置错误:next值必须为时间格式yyyy/mm/dd HH:mm:ss（%s,%s）(%v)", name, c.String("next"), err)
			}
			if name == "" {
				return fmt.Errorf("task配置错误:name不能为空（name:%s）", name)
			}
			tasks = append(tasks, NewTask(name,
				time.Duration(next.Sub(time.Now()).Seconds()),
				time.Duration(interval), w.handle(service, method, params), fmt.Sprintf("%s-%s", service, method)))
		}
		for _, task := range tasks {
			w.server.Add(task)
		}
	}
	//设置metric上报
	metric, err := conf.GetNode("metric")
	if r, ok := v.conf.GetNode("metric"); ok && r.GetVersion() != metric.GetVersion() || !ok {
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
	//设置基本参数
	w.server.SetName(conf.String("name", "cron.server"))
	w.conf = conf
	return nil

}

//setRouter 设置路由
func (w *hydraCronServer) handle(service, method, params string) func(task *Task) error {
	return func(task *Task) error {
		//处理输入参数
		hydraContext := make(map[string]interface{})

		//执行服务调用
		response, err := w.handler.Handle(task.taskName, method, service, params, hydraContext)
		if err != nil {
			return err
		}
		if response.Status == 0 {
			response.Status = 200
		}
		task.Result = response.Content
		task.statusCode = response.Status
		return nil
	}
}

//GetAddress 获取服务器地址
func (w *hydraCronServer) GetAddress() string {
	return w.server.ip
}

//Start 启用服务
func (w *hydraCronServer) Start() (err error) {
	w.server.Start()
	return nil
}

//接口服务变更通知
func (w *hydraCronServer) Notify(conf registry.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil && w.conf.GetVersion() == conf.GetVersion() {
		return errors.New("版本无变化")
	}
	return w.restartServer(conf)

}

//Shutdown 关闭服务
func (w *hydraCronServer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (server.IHydraServer, error) {
	return newHydraCronServer(c, r, conf, logger)
}

func init() {
	server.Register("cron.server", &hydraCronServerAdapter{})
}
