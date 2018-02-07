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

	//设置server节点
	serverRaw, err := cnf.GetServerRaw()
	if err != nil {
		return fmt.Errorf("%s:server配置有误:%v", cnf.GetFullName(), err)
	}

	//设置queue
	queues, err := cnf.GetQueues()
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		err = fmt.Errorf("%s:queue配置有误:%v", cnf.GetFullName(), err)
		return err
	}
	if err != conf.ERR_NO_CHANGED {
		ext := map[string]interface{}{
			"__get_sharding_index_": func() (int, int) {
				return w.shardingIndex, w.shardingCount
			},
		}
		for _, queue := range queues {
			queue.Handler = middleware.ContextHandler(w.engine, queue.Name, queue.Engine, queue.Service, queue.Setting, ext)
		}
		err = w.server.SetQueues(serverRaw, queues...)
		if err != nil {
			return fmt.Errorf("queue配置有误:%v", err)
		}
		w.Infof("%s:queue配置:%d", cnf.GetFullName(), len(queues))
	}

	if err == nil && len(queues) == 0 {
		w.Infof("%s:未配置queue", cnf.GetFullName())
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
