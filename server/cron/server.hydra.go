package cron

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web server适配器
type hydraCronServer struct {
	server   *CronServer
	conf     conf.Conf
	registry context.IServiceRegistry
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraCronServer(handler context.EngineHandler, r context.IServiceRegistry, cnf conf.Conf) (h *hydraCronServer, err error) {
	h = &hydraCronServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		server: NewCronServer(cnf.String("name", "cron.server"),
			60,
			time.Second,
			WithRegistry(r),
			WithIP(net.GetLocalIPAddress(cnf.String("mask")))),
	}
	err = h.setConf(cnf)
	return
}

//restartServer 重启服务器
func (w *hydraCronServer) restartServer(conf conf.Conf) (err error) {
	w.Shutdown()
	w.server = NewCronServer(conf.String("name", "cron.server"),
		60,
		time.Second,
		WithRegistry(w.registry),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	err = w.setConf(conf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraCronServer) setConf(conf conf.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置任务
	routers, err := conf.GetNode("task")
	if err != nil {
		return fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNode("task"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("tasks")
		if err != nil {
			return fmt.Errorf("tasks未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		tasks := make([]*Task, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			method := c.String("method")
			params := c.String("params")
			interval, err := time.ParseDuration(c.String("interval", "-1"))
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
	if err != nil {
		return fmt.Errorf("metric未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNode("metric"); err != nil || r.GetVersion() != metric.GetVersion() {
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
func (w *hydraCronServer) handle(service, method, args string) func(task *Task) error {
	return func(task *Task) (err error) {
		//处理输入参数
		context := context.GetContext()
		defer context.Close()
		context.Ext["hydra_sid"] = task.GetSessionID()
		context.Input.Args, err = utility.GetMapWithQuery(args)
		if err != nil {
			task.statusCode = 500
			task.Result = err
			return err
		}
		//执行服务调用
		response, err := w.handler.Handle(task.taskName, method, service, context)
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
func (w *hydraCronServer) Notify(conf conf.Conf) error {
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

func (h *hydraCronServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraCronServer(c, r, conf)
}

func init() {
	server.Register("cron", &hydraCronServerAdapter{})
}
