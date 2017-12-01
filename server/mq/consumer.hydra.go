package mq

import (
	"fmt"
	"reflect"
	"strings"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

//hydraWebServer mq consumer服务器
type hydraMQConsumer struct {
	domain   string
	server   *MQConsumer
	registry server.IServiceRegistry
	conf     conf.Conf
	handler  server.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建mq consumer服务器
func newHydraMQConsumer(handler server.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraMQConsumer, err error) {
	h = &hydraMQConsumer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
	}
	h.server, err = NewMQConsumer(cnf.String("domain"), cnf.String("name", "mq.server"),
		cnf.String("address"),
		WithVersion(cnf.String("version")),
		WithRaw(cnf.GetContent()),
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
	w.server, err = NewMQConsumer(cnf.String("domain"), cnf.String("name", "mq.server"),
		cnf.String("address"),
		WithVersion(cnf.String("version")),
		WithRaw(cnf.GetContent()),
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
	//检查版本是否一致
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查状态是否一致
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}
	//设置监控的对列
	queues, err := server.GetQueues(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED {
		err = fmt.Errorf("queue配置有误:%v", err)
		return err
	}
	if err == nil {
		for _, queue := range queues {
			err := w.server.Use(queue.Name, w.handle(queue.Service, queue.Mode, "", queue.Args))
			if err != nil {
				return err
			}
		}
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
	//更新配置
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
		data := make(map[string]interface{})
		body := task.msg.GetMessage()
		if strings.HasPrefix(body, "{") && strings.HasSuffix(body, "}") {
			data, err = jsons.Unmarshal([]byte(body))
			if err != nil {
				task.statusCode = 500
				task.Result = fmt.Errorf("输入参数不是有效的json字符串:%s", body)
				return err
			}
		}
		margs, err := utility.GetMapWithQuery(args)
		if err != nil {
			task.statusCode = 500
			task.Result = fmt.Errorf("args格式错误:%s(err:%v)", args, err)
			return err
		}

		input := transform.NewMaps(data).Data
		params := transform.NewMaps(make(map[string]interface{})).Data
		ext := make(map[string]interface{})
		ext["hydra_sid"] = task.GetSessionID()
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#@domain/var/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return string(cnf), nil
		}

		//执行服务调用
		ctx.SetInput(input, params, body, margs, ext)
		response, err := w.handler.Handle(task.queue, mode, service, ctx)
		if reflect.ValueOf(response).IsNil() {
			response = context.GetStandardResponse()
		}
		defer func() {
			response.Close()
			if err != nil {
				task.Errorf("mq.response.error: %v", task.err)
			}
		}()
		if err != nil {
			task.err = fmt.Errorf("mq.server.handler.error:%s,%v,%v", task.queue, response.GetContent(), err)
			task.statusCode = response.GetStatus(task.err)
			task.Result = response.GetContent(task.err)
			return task.err
		}
		task.statusCode = response.GetStatus()
		task.Result = response.GetContent()
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
	if !strings.EqualFold(conf.String("version"), w.conf.String("version")) {
		return true, nil
	}
	routers, err := conf.GetNodeWithSectionName("queue", "#@path/queue")
	if err != nil {
		return false, fmt.Errorf("queue未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := conf.GetNodeWithSectionName("queue", "#@path/queue"); err != nil || r.GetVersion() != routers.GetVersion() {
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
func (w *hydraMQConsumer) GetServices() []string {
	return w.handler.GetServices()
}

//Shutdown 关闭服务
func (w *hydraMQConsumer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c server.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraMQConsumer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_MQ, &hydraCronServerAdapter{})
}
