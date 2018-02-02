package registry

import (
	"fmt"

	"github.com/qxnw/hydra/servers/http/middleware"
)

//SetConf 设置配置参数
func (w *RegistryServer) SetConf(conf *RegistryConf) error {
	//检查版本号
	if !conf.IsChanged() {
		return nil
	}
	//检查服务器状态
	if conf.IsStoped() {
		return fmt.Errorf("%s:配置为:stop", conf.GetFullName())
	}

	//设置路由
	routers, err := conf.GetRouters()
	if err != nil && err != ERR_NO_CHANGED && err != ERR_NOT_SETTING {
		err = fmt.Errorf("%s:路由配置有误:%v", conf.GetFullName(), err)
		return err
	}
	if err != ERR_NO_CHANGED {
		for _, router := range routers {
			router.Handler = middleware.ContextHandler(w.engine, router.Name, router.Engine, router.Service, router.Setting)
		}
		err = w.server.SetRouters(routers)
		if err != nil {
			return fmt.Errorf("路由配置有误:%v", err)
		}
		w.Infof("%s:路由配置:%d", conf.GetFullName(), len(routers))
	}

	//设置静态文件路由
	enable, prefix, dir, showDir, exts, err1 := conf.GetStatic()
	if err1 != nil && err1 != ERR_NO_CHANGED && err1 != ERR_NOT_SETTING {
		return err1
	}
	if err1 == ERR_NOT_SETTING || !enable {
		w.Infof("%s:未配置静态文件", conf.GetFullName())
		w.server.SetStatic(false, prefix, dir, showDir, exts)
	}
	if err1 == nil && enable {
		w.Infof("%s:启用静态文件", conf.GetFullName())
		w.server.SetStatic(true, prefix, dir, showDir, exts)
	}
	if err == ERR_NOT_SETTING && err1 == ERR_NOT_SETTING {
		return fmt.Errorf("路由配置有误:%v，静态文件:%v", err, err1)
	}
	if err != nil && err1 != nil && err != ERR_NO_CHANGED && err1 != ERR_NO_CHANGED {
		return fmt.Errorf("路由配置有误:%v，静态文件:%v", err, err1)
	}
	if err == nil && len(routers) == 0 {
		w.Infof("%s:未配置路由", conf.GetFullName())
	}
	//设置通用头信息
	headers, err := conf.GetHeaders()
	if err != nil && err != ERR_NO_CHANGED && err != ERR_NOT_SETTING {
		return err
	}
	if err == nil || err == ERR_NOT_SETTING {
		w.Infof("%s:http header:%d", conf.GetFullName(), len(headers))
		w.server.SetHeader(headers)
	}

	//设置jwt安全认证参数
	jwt, err := conf.GetAuth("jwt")
	if err != nil && err != ERR_NO_CHANGED && err != ERR_NOT_SETTING {
		return err
	}
	if err == ERR_NOT_SETTING || !jwt.Enable {
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
	}
	if err == nil && jwt.Enable {
		w.Infof("%s:启用jwt校验", conf.GetFullName())
		w.server.SetJWT(jwt.Enable, jwt.Name, jwt.Mode, jwt.Secret, jwt.Exclude, jwt.ExpireAt)
	}
	//设置OnlyAllowAjaxRequest
	enable = conf.GetOnlyAllowAjaxRequest()
	w.server.SetOnlyAllowAjaxRequest(enable)
	if enable {
		w.Infof("%s:启用ajax调用限制", conf.GetFullName())
	}

	//设置metric服务器监控数据
	enable, host, dataBase, userName, password, span, err := conf.GetMetric()
	if err != nil && err != ERR_NO_CHANGED && err != ERR_NOT_SETTING {
		w.Errorf("%s:metric配置有误(%v)", conf.GetFullName(), err)
		w.server.StopMetric()
	}
	if err == ERR_NOT_SETTING || !enable {
		w.Warnf("%s:未配置metric", conf.GetFullName())
		w.server.StopMetric()
	}
	if err == nil && enable {
		w.server.Infof("%s:启用metric", conf.GetFullName())
		w.server.SetMetric(host, dataBase, userName, password, span)
	}

	//设置其它参数
	w.server.SetHost(conf.GetHost())
	return nil
}
