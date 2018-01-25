package cron

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"sync"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
	"github.com/zkfy/cron"
)

//hydraWebServer web server适配器
type hydraCronServer struct {
	server   *CronServer
	conf     conf.Conf
	registry server.IServiceRegistry
	handler  server.EngineHandler
	log      *logger.Logger
	mu       sync.Mutex
}

//newHydraRPCServer 构建基本配置参数的web server
func newHydraCronServer(handler server.EngineHandler, r server.IServiceRegistry, cnf conf.Conf, log *logger.Logger) (h *hydraCronServer, err error) {
	h = &hydraCronServer{handler: handler,
		conf:     conf.NewJSONConfWithEmpty(),
		registry: r,
		log:      log,
		server: NewCronServer(cnf.String("domain"), cnf.String("name", "cron.server"),
			60,
			time.Second,
			WithRegistry(r, cnf.Translate("{@category_path}/servers/{@tag}")),
			WithIP(net.GetLocalIPAddress(cnf.String("mask"))), WithLogger(log)),
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
		WithIP(net.GetLocalIPAddress(cnf.String("mask"))), WithLogger(w.log))
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
	//设置cron任务
	tasks, err := server.GetTasks(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED {
		err = fmt.Errorf("task配置有误:%v", err)
		return err
	}
	if err == nil {
		for _, task := range tasks {
			s, err := cron.ParseStandard(task.Cron)
			if err != nil {
				return fmt.Errorf("task的cron未配置或配置有误:%s(cron:%s,err:%+v)", conf.String("name"), task.Cron, err)
			}
			tk := NewTask(task.Name, s, w.handle(task), task.Service)
			w.server.Add(tk)
		}
	}

	//设置metric服务器监控数据
	enable, host, dataBase, userName, password, span, err := server.GetMetric(w.conf, conf)
	if err != nil && err != server.ERR_NO_CHANGED && err != server.ERR_NOT_SETTING {
		w.server.Errorf("%s.%s:metric配置有误(%v)", conf.String("name"), conf.String("type"), err)
		w.server.StopInfluxMetric()
	}
	if err == server.ERR_NOT_SETTING || !enable {
		w.server.Warnf("%s.%s:未配置metric", conf.String("name"), conf.String("type"))
		w.server.StopInfluxMetric()
	}
	if err == nil && enable {
		w.server.Infof("%s.%s:启用metric", conf.String("name"), conf.String("type"))
		w.server.SetInfluxMetric(host, dataBase, userName, password, span)
	}
	w.server.cluster = conf.String("cluster", "master-slave")
	w.server.Infof("%s.%s启动模式:%s", conf.String("name"), conf.String("type"), w.server.cluster)
	if len(conf.String("rds-records")) > 0 {
		w.server.Infof("%s.%s:启用自动保存执行记录", conf.String("name"), conf.String("type"))
	} else {
		w.server.Warnf("%s.%s:未启用自动保存执行记录", conf.String("name"), conf.String("type"))
	}
	//设置基本参数
	w.conf = conf
	return nil

}

//setRouter 执行引擎操作
func (w *hydraCronServer) handle(xtask *server.Task) func(task *Task) error {
	return func(task *Task) (err error) {
		//处理输入参数
		ctx := context.GetContext()
		defer ctx.Close()

		ext := map[string]interface{}{"hydra_sid": task.GetSessionID()}
		var inputGetter transform.ITransformGetter
		var paramGetter transform.ITransformGetter
		var inputBody string
		if xtask.Input != "" {
			input, err := utility.GetMapWithQuery(xtask.Input)
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
		if xtask.Body != "" {
			if strings.HasPrefix(xtask.Body, "#") {
				cnf, err := w.conf.GetRawNodeWithValue(xtask.Body, true)
				if err != nil {
					task.statusCode = 500
					task.err = err
					task.Result = err
					task.Errorf("获取body节点(%s)数据失败:(err:%v)", xtask.Body, task.err)
					return err
				}
				inputBody = string(cnf)
			} else {
				inputBody = xtask.Body
			}
		}
		ext["__func_body_get_"] = func(ch string) (string, error) {
			return inputBody, nil
		}
		ext["__func_var_get_"] = func(c string, n string) (string, error) {
			cnf, err := w.conf.GetRawNodeWithValue(fmt.Sprintf("#/@domain/var/%s/%s", c, n), false)
			if err != nil {
				return "", err
			}
			return string(cnf), nil
		}
		ext["__cron_"] = xtask.Cron
		margs, err := utility.GetMapWithQuery(xtask.Args)
		if err != nil {
			task.statusCode = 500
			task.err = fmt.Errorf("args配置出错(%s)：%v", xtask.Args, err)
			task.Result = task.err
			task.Error(task.err)
			return
		}
		//执行服务调用
		ctx.SetInput(inputGetter, paramGetter, margs, ext)
		response, err := w.handler.Execute(task.taskName, xtask.Mode, xtask.Service, ctx)
		if reflect.ValueOf(response).IsNil() {
			response = context.GetStandardResponse()
		}
		defer func() {
			response.Close()
			if err != nil {
				task.Errorf("cron.response.error: %v", task.err)
			}
			xtask.Result = response.GetStatus()
			w.saveHistory2Redis(xtask)
		}()
		if err != nil {
			task.err = fmt.Errorf("cron.server.handler.error:%v,%v", response.GetContent(), err)
			task.statusCode = response.GetStatus(task.err)
			task.Result = response.GetContent()
			return task.err
		}
		task.Result = response.GetContent()
		task.statusCode = response.GetStatus()
		return nil
	}
}
func (w *hydraCronServer) saveHistory2Redis(xtask *server.Task) {
	if len(w.conf.String("rds-records")) > 0 {
		s, _ := cron.ParseStandard(xtask.Cron)
		xtask.Next = s.Next(time.Now()).Format("20060102150405")
		xtask.Last = time.Now().Format("20060102150405")
		err := server.SaveCronExecuteHistory(w.conf.String("rds-records"), w.conf, xtask)
		if err != nil {
			w.server.Errorf("保存cron的执行记录失败:%v", err)
		}
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
	if conf.String("cluster") != w.conf.String("cluster") {
		return true, nil
	}
	routers, err := conf.GetNodeWithSectionName("task", "#@path/task")
	if err != nil {
		return false, fmt.Errorf("task未配置或配置有误:%s(%+v)", conf.String("name"), err)
	}
	//检查路由是否变化，已变化则需要重启服务
	if r, err := w.conf.GetNodeWithSectionName("task", "#@path/task"); err != nil || r.GetVersion() != routers.GetVersion() {
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
func (w *hydraCronServer) GetServices() []string {
	return w.handler.GetServices()
}

//Shutdown 关闭服务
func (w *hydraCronServer) Shutdown() {
	w.server.Close()
}

type hydraCronServerAdapter struct {
}

func (h *hydraCronServerAdapter) Resolve(c server.EngineHandler, r server.IServiceRegistry, conf conf.Conf, log *logger.Logger) (server.IHydraServer, error) {
	return newHydraCronServer(c, r, conf, log)
}

func init() {
	server.Register(server.SRV_TP_CRON, &hydraCronServerAdapter{})
}
