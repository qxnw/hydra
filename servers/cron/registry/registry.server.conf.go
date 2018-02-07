package registry

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
)

//SetConf 设置配置参数
func (w *RegistryServer) SetConf(cnf *conf.RegistryConf) error {
	//检查版本号
	if !cnf.IsChanged() {
		return nil
	}
	//检查服务器状态
	if cnf.IsStoped() {
		return fmt.Errorf("%s:配置为:stop", cnf.GetFullName())
	}
	redisConf, _ := cnf.GetRedisRaw()
	//设置tasks
	tasks, err := cnf.GetTasks()
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		err = fmt.Errorf("%s:queue配置有误:%v", cnf.GetFullName(), err)
		return err
	}
	if err != conf.ERR_NO_CHANGED {
		for _, task := range tasks {
			ext := map[string]interface{}{
				"__get_sharding_index_": func() (int, int) {
					return w.shardingIndex, w.shardingCount
				},
			}
			ext["__cron_"] = task.Cron
			task.Handler = middleware.ContextHandler(w.engine, task.Name, task.Engine, task.Service, task.Setting, ext)
		}
		err = w.server.SetTasks(redisConf, tasks)
		if err != nil {
			return fmt.Errorf("task配置有误:%v", err)
		}
		w.Infof("%s:task配置:%d", cnf.GetFullName(), len(tasks))
	}

	if err == nil && len(tasks) == 0 {
		w.Infof("%s:未配置task", cnf.GetFullName())
	}

	//设置metric服务器监控数据
	enable, host, dataBase, userName, password, span, err := cnf.GetMetric()
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		w.Errorf("%s:metric配置有误(%v)", cnf.GetFullName(), err)
		w.server.StopMetric()
	}
	if err == conf.ERR_NOT_SETTING || !enable {
		w.Warnf("%s:未配置metric", cnf.GetFullName())
		w.server.StopMetric()
	}
	if err == nil && enable {
		w.server.Infof("%s:启用metric", cnf.GetFullName())
		w.server.SetMetric(host, dataBase, userName, password, span)
	}

	return nil
}
