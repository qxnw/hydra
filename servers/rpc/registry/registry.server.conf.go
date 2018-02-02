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

	//设置路由
	routers, err := cnf.GetRouters()
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		err = fmt.Errorf("%s:路由配置有误:%v", cnf.GetFullName(), err)
		return err
	}
	if err != conf.ERR_NO_CHANGED {
		for _, router := range routers {
			router.Handler = middleware.ContextHandler(w.engine, router.Name, router.Engine, router.Service, router.Setting)
		}
		err = w.server.SetRouters(routers)
		if err != nil {
			return fmt.Errorf("路由配置有误:%v", err)
		}
		w.Infof("%s:路由配置:%d", cnf.GetFullName(), len(routers))
	}

	if err == nil && len(routers) == 0 {
		w.Infof("%s:未配置路由", cnf.GetFullName())
	}
	//设置通用头信息
	headers, err := cnf.GetHeaders()
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		return err
	}
	if err == nil || err == conf.ERR_NOT_SETTING {
		w.Infof("%s:http header:%d", cnf.GetFullName(), len(headers))
		w.server.SetHeader(headers)
	}

	//设置jwt安全认证参数
	jwt, err := cnf.GetAuth("jwt")
	if err != nil && err != conf.ERR_NO_CHANGED && err != conf.ERR_NOT_SETTING {
		return err
	}
	if err == conf.ERR_NOT_SETTING || !jwt.Enable {
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
	}
	if err == nil && jwt.Enable {
		w.Infof("%s:启用jwt校验", cnf.GetFullName())
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
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

	//设置其它参数
	w.server.SetHost(cnf.GetHost())
	return nil
}
