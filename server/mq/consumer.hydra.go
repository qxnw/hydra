package mq

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
type hydraMQConsumer struct {
	server   *MQConsumer
	logger   context.Logger
	registry context.IServiceRegistry
	conf     registry.Conf
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraMQConsumer(handler context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (h *hydraMQConsumer, err error) {
	h = &hydraMQConsumer{handler: handler,
		logger:   logger,
		conf:     registry.NewJSONConfWithEmpty(),
		registry: r,
	}
	h.server, err = NewMQConsumer(conf.String("name", "mq.consumer"),
		conf.String("address"),
		conf.String("version"),
		WithLogger(logger),
		WithRegistry(r),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	if err != nil {
		return
	}
	err = h.setConf(conf)
	return
}

//restartServer 重启服务器
func (w *hydraMQConsumer) restartServer(conf registry.Conf) (err error) {
	w.Shutdown()
	w.server, err = NewMQConsumer(conf.String("name", "mq.consumer"),
		conf.String("address"),
		conf.String("version"),
		WithLogger(w.logger),
		WithRegistry(w.registry),
		WithIP(net.GetLocalIPAddress(conf.String("mask"))))
	if err != nil {
		return
	}
	err = w.setConf(conf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraMQConsumer) setConf(conf registry.Conf) error {
	if w.conf.GetVersion() == conf.GetVersion() {
		return fmt.Errorf("配置版本无变化(%s,%d)", w.server.serverName, w.conf.GetVersion())
	}
	//设置路由
	routers, err := conf.GetNode("queue")
	if err != nil {
		return fmt.Errorf("queue未配置或配置有误:%s(err:%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNode("queue"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("queues")
		if err != nil {
			return fmt.Errorf("queues未配置或配置有误:%s(err:%+v)", conf.String("name"), err)
		}
		queues := make([]task, 0, len(rts))
		for _, c := range rts {
			queue := c.String("name")
			service := c.String("service")
			method := c.String("method")
			params := c.String("params")
			if queue == "" || service == "" || method == "" {
				return fmt.Errorf("queue配置错误:service 和 name不能为空（name:%s，service:%s）", queue, service)
			}
			queues = append(queues, task{name: queue, service: service, method: method, params: params})
		}
		for _, task := range queues {
			go w.server.Use(task.name, w.handle(task.service, task.method, task.params))
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
	w.conf = conf
	return nil

}

//setRouter 设置路由
func (w *hydraMQConsumer) handle(service, method, params string) func(task *Context) error {
	return func(task *Context) error {
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
func (w *hydraMQConsumer) GetAddress() string {
	return w.server.ip
}

//Start 启用服务
func (w *hydraMQConsumer) Start() (err error) {
	return w.server.Run()
}

//接口服务变更通知
func (w *hydraMQConsumer) Notify(conf registry.Conf) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conf != nil && w.conf.GetVersion() == conf.GetVersion() {
		return errors.New("版本无变化")
	}
	return w.restartServer(conf)
}

//Shutdown 关闭服务
func (w *hydraMQConsumer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c context.EngineHandler, r context.IServiceRegistry, conf registry.Conf, logger context.Logger) (server.IHydraServer, error) {
	return newHydraMQConsumer(c, r, conf, logger)
}

func init() {
	server.Register("mq.consumer", &hydraCronServerAdapter{})
}
