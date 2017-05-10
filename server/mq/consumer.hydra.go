package mq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer web server适配器
type hydraMQConsumer struct {
	server   *MQConsumer
	registry server.IServiceRegistry
	conf     conf.Conf
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraMQConsumer(handler context.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraMQConsumer, err error) {
	h = &hydraMQConsumer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
	}
	h.server, err = NewMQConsumer(cnf.String("name", "mq.server"),
		cnf.String("address"),
		WithVersion(cnf.String("version")),
		WithRegistry(r, cnf.Translate("{@category_path}/servers/{@tag}")),
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	if err != nil {
		return
	}
	err = h.setConf(cnf)
	return
}

//restartServer 重启服务器
func (w *hydraMQConsumer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	w.server, err = NewMQConsumer(cnf.String("name", "mq.server"),
		cnf.String("address"),
		WithVersion(cnf.String("version")),
		WithRegistry(w.registry, cnf.Translate("{@category_path}/servers/{@tag}")),
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	if err != nil {
		return
	}
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraMQConsumer) setConf(conf conf.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}
	//设置消息队列
	routers, err := conf.GetNodeWithSection("queue")
	if err != nil {
		return fmt.Errorf("queue未配置或配置有误:%s(err:%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSection("queue"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("queues")
		if err != nil {
			return fmt.Errorf("queues未配置或配置有误:%s(err:%+v)", conf.String("name"), err)
		}
		queues := make([]task, 0, len(rts))
		for _, c := range rts {
			queue := c.String("name")
			service := c.String("service")
			action := c.String("action")
			mode := c.String("mode", "*")
			args := c.String("args")
			if queue == "" || service == "" || action == "" {
				return fmt.Errorf("queue配置错误:name,service,action不能为空（name:%s，service:%s，action:%s）", queue, service, action)
			}
			queues = append(queues, task{name: queue, service: service, action: action, args: args, mode: mode})
		}
		for _, task := range queues {
			w.server.Use(task.name, w.handle(task.service, task.mode, task.action, task.args))
		}

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
			w.server.SetInfluxMetric(host, dataBase, userName, password, 10*time.Second)
		}
	} else {
		w.server.StopInfluxMetric()
	}

	w.conf = conf
	return nil

}

//setRouter 设置路由
func (w *hydraMQConsumer) handle(service, mode, method, args string) func(task *Context) error {
	return func(task *Context) error {
		//处理输入参数
		var err error
		ctx := context.GetContext()
		defer ctx.Close()
		ctx.Input.Input = json.RawMessage(task.params)
		ctx.Input.Args, err = utility.GetMapWithQuery(args)
		if err != nil {
			task.statusCode = 500
			task.Result = err
			return err
		}
		ctx.Ext["hydra_sid"] = task.GetSessionID()
		ctx.Ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetNodeWithValue(fmt.Sprintf("#@domain/var/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return cnf.GetContent(), nil
		}

		//执行服务调用
		start := time.Now()
		response, err := w.handler.Handle(task.queue, mode, service, ctx)
		if err != nil {
			task.statusCode = 500
			task.err = err
			if server.IsDebug {
				task.Errorf("mq:%s(%v),err:%v", task.queue, time.Since(start), task.err)
				return err
			}
			task.err = errors.New("Internal Server Error(工作引擎发生异常)")
			task.Errorf("mq:%s(%v),err:%v", task.queue, time.Since(start), task.err)
			return task.err
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
func (w *hydraMQConsumer) GetAddress() string {
	return w.server.ip
}

//Start 启用服务
func (w *hydraMQConsumer) Start() (err error) {
	return w.server.Run()
}

//接口服务变更通知
func (w *hydraMQConsumer) Notify(conf conf.Conf) error {
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
func (w *hydraMQConsumer) needRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	routers, err := conf.GetNodeWithSection("queue")
	if err != nil {
		return false, fmt.Errorf("queue未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSection("queue"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	return false, nil
}
func (w *hydraMQConsumer) GetStatus() string {
	if w.server.running {
		return server.ST_RUNNING
	}
	return server.ST_STOP
}

//Shutdown 关闭服务
func (w *hydraMQConsumer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c context.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraMQConsumer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_MQ, &hydraCronServerAdapter{})
}
