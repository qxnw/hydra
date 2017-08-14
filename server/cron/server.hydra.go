package cron

import (
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
	"github.com/qxnw/lib4go/utility"
	"github.com/zkfy/cron"
)

//hydraWebServer web server适配器
type hydraCronServer struct {
	server   *CronServer
	conf     conf.Conf
	registry server.IServiceRegistry
	handler  context.EngineHandler
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraCronServer(handler context.EngineHandler, r server.IServiceRegistry, cnf conf.Conf) (h *hydraCronServer, err error) {
	h = &hydraCronServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		server: NewCronServer(cnf.String("domain"), cnf.String("name", "cron.server"),
			60,
			time.Second,
			WithRegistry(r, cnf.Translate("{@category_path}/servers/{@tag}")),
			WithIP(net.GetLocalIPAddress(cnf.String("mask")))),
	}
	err = h.setConf(cnf)
	return
}

//restartServer 重启服务器
func (w *hydraCronServer) restartServer(cnf conf.Conf) (err error) {
	w.Shutdown()
	w.server = NewCronServer(cnf.String("domain"), cnf.String("name", "cron.server"),
		60,
		time.Second,
		WithRegistry(w.registry, cnf.Translate("{@category_path}/servers/{@tag}")),
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))))
	w.conf = conf.NewJSONConfWithEmpty()
	err = w.setConf(cnf)
	if err != nil {
		return
	}
	return w.Start()
}

//SetConf 设置配置参数
func (w *hydraCronServer) setConf(conf conf.Conf) error {
	//检查配置版本是否变更
	if w.conf.GetVersion() == conf.GetVersion() {
		return nil
	}
	//检查服务器状态是否停止
	if strings.EqualFold(conf.String("status"), server.ST_STOP) {
		return fmt.Errorf("服务器配置为:%s", conf.String("status"))
	}
	//设置任务
	routers, err := conf.GetNodeWithSectionName("task")
	if err != nil {
		return fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	if r, err := w.conf.GetNodeWithSectionName("task"); err != nil || r.GetVersion() != routers.GetVersion() {
		baseArgs := routers.String("args")
		rts, err := routers.GetSections("tasks")
		if err != nil {
			return fmt.Errorf("tasks未配置或配置有误:%s(%+v)", conf.String("name"), err)
		}
		tasks := make([]*Task, 0, len(rts))
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			input := c.String("input")
			body := c.String("body")
			args := c.String("args")
			mode := c.String("mode", "*")
			cronStr := c.String("cron")
			if name == "" || service == "" || cronStr == "" {
				return fmt.Errorf("task配置错误:name,service,cron不能为空（name:%s，service:%s,cron:%s）", name, service, cronStr)
			}

			s, err := cron.ParseStandard(cronStr)
			if err != nil {
				return fmt.Errorf("task的cron未配置或配置有误:%s(cron:%s,err:%+v)", conf.String("name"), cronStr, err)
			}
			tasks = append(tasks, NewTask(name, s, w.handle(service, mode, input, body, baseArgs+"&"+args), service))
		}
		for _, task := range tasks {
			w.server.Add(task)
		}
	}
	//设置metric上报监控数据
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

	//设置基本参数
	w.conf = conf
	return nil

}

//setRouter 执行引擎操作
func (w *hydraCronServer) handle(service, mode, input, body, args string) func(task *Task) error {
	return func(task *Task) (err error) {
		//处理输入参数
		ctx := context.GetContext()
		defer ctx.Close()

		ext := map[string]interface{}{"hydra_sid": task.GetSessionID()}
		var inputGetter transform.ITransformGetter
		var paramGetter transform.ITransformGetter
		var inputBody string
		if input != "" {
			input, err := utility.GetMapWithQuery(input)
			if err != nil {
				task.statusCode = 500
				task.Result = err
				return err
			}
			inputGetter = transform.NewMap(input).Data
			paramGetter = inputGetter
		} else {
			inputGetter = transform.NewMap(make(map[string]string)).Data
			paramGetter = inputGetter
		}
		if body != "" {
			if strings.HasPrefix(body, "#") {
				cnf, err := w.conf.GetRawNodeWithValue(body, true)
				if err != nil {
					task.statusCode = 500
					task.err = err
					task.Result = err
					task.Errorf("获取body节点(%s)数据失败:(err:%v)", body, task.err)
					return err
				}
				inputBody = string(cnf)
			} else {
				inputBody = body
			}
		}
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#@domain/var/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return string(cnf), nil
		}
		margs, err := utility.GetMapWithQuery(args)
		if err != nil {
			task.statusCode = 500
			task.err = fmt.Errorf("args配置出错(%s)：%v", args, err)
			task.Result = task.err
			task.Error(task.err)
			return
		}
		//执行服务调用
		ctx.SetInput(inputGetter, paramGetter, inputBody, margs, ext)
		response, err := w.handler.Handle(task.taskName, mode, service, ctx)
		if response == nil {
			response = &context.Response{}
		}
		defer func() {
			if err != nil {
				task.Errorf("cron.response.error: %v", task.err)
			}
		}()
		if err != nil || (response.Status >= 500 && response.Status < 600) {
			task.err = fmt.Errorf("cron.server.handler.error:%v,%v", response.Content, err)
			response.Status = types.DecodeInt(response.Status, 0, 500, response.Status)
			task.statusCode = response.Status
			response.Content = task.err.Error()
			return task.err
		}
		response.Status = types.DecodeInt(response.Status, 0, 200, response.Status)
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
	return w.server.Start()
}

//接口服务变更通知
func (w *hydraCronServer) Notify(conf conf.Conf) error {
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

//needRestart 检查是否需要重启
func (w *hydraCronServer) needRestart(conf conf.Conf) (bool, error) {
	if !strings.EqualFold(conf.String("status"), w.conf.String("status")) {
		return true, nil
	}
	routers, err := conf.GetNodeWithSectionName("task")
	if err != nil {
		return false, fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("task"); err != nil || r.GetVersion() != routers.GetVersion() {
		return true, nil
	}
	return false, nil
}

//GetStatus 获取服务器运行状态
func (w *hydraCronServer) GetStatus() string {
	if w.server.running {
		return server.ST_RUNNING
	}
	return server.ST_STOP
}

//Shutdown 关闭服务
func (w *hydraCronServer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c context.EngineHandler, r server.IServiceRegistry, conf conf.Conf) (server.IHydraServer, error) {
	return newHydraCronServer(c, r, conf)
}

func init() {
	server.Register(server.SRV_TP_CRON, &hydraCronServerAdapter{})
}
