package cron

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/middleware"
)

type ISetMetric interface {
	SetMetric(*conf.Metric) error
}

//SetMetric 设置metric
func SetMetric(set ISetMetric, cnf conf.IServerConf) (enable bool, err error) {
	//设置静态文件路由
	var metric conf.Metric
	_, err = cnf.GetSubObject("metric", &metric)
	if err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if b, err := govalidator.ValidateStruct(&metric); !b {
		err = fmt.Errorf("metric配置有误:%v", err)
		return false, err
	}
	err = set.SetMetric(&metric)
	return enable, err
}

//ITasks 设置tasks
type ITasks interface {
	SetTasks(string, []*conf.Task) error
}

//SetTasks 设置tasks
func SetTasks(engine servers.IExecuter, set ITasks, cnf conf.IServerConf, ext map[string]interface{}) (enable bool, err error) {
	reidsConf, err := cnf.GetSubConf("redis")
	if err != nil && err != conf.ErrNoSetting {
		return false, err
	}

	var tasks conf.Tasks
	if _, err = cnf.GetSubObject("task", &tasks); err == conf.ErrNoSetting {
		err = fmt.Errorf("task:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}
	if b, err := govalidator.ValidateStruct(&tasks); !b {
		err = fmt.Errorf("task配置有误:%v", err)
		return false, err
	}
	for _, task := range tasks.Tasks {
		task.Handler = middleware.ContextHandler(engine, task.Name, task.Engine, task.Service, task.Setting, ext)
	}
	if err = set.SetTasks(string(reidsConf.GetRaw()), tasks.Tasks); err != nil {
		return false, err
	}
	return len(tasks.Tasks) > 0, nil
}
