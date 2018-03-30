package mqc

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

//IQueues 设置queue
type IQueues interface {
	SetQueues(string, []*conf.Queue) error
}

//SetQueues 设置queue
func SetQueues(engine servers.IRegistryEngine, set IQueues, cnf conf.IServerConf, ext map[string]interface{}) (enable bool, err error) {

	serverConf, err := cnf.GetSubConf("server")
	if err == conf.ErrNoSetting {
		err = fmt.Errorf("server节点:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}

	var queues conf.Queues
	if _, err = cnf.GetSubObject("queue", &queues); err == conf.ErrNoSetting {
		err = fmt.Errorf("queue:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}
	if b, err := govalidator.ValidateStruct(&queues); !b {
		err = fmt.Errorf("queue配置有误:%v", err)
		return false, err
	}
	for _, queue := range queues.Queues {
		queue.Handler = middleware.ContextHandler(engine, engine, queue.Name, queue.Engine, queue.Service, queue.Setting, ext)
	}
	if err = set.SetQueues(string(serverConf.GetRaw()), queues.Queues); err != nil {
		return false, err
	}
	return len(queues.Queues) > 0, nil
}
