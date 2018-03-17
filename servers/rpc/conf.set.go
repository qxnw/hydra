package rpc

import (
	"fmt"

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
	err = set.SetMetric(&metric)
	return enable, err
}

type ISetStatic interface {
	SetStatic(static *conf.Static) error
}

//SetStatic 设置static
func SetStatic(set ISetStatic, cnf conf.IServerConf) (enable bool, err error) {
	//设置静态文件路由
	var static conf.Static
	_, err = cnf.GetSubObject("static", &static)
	if err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = set.SetStatic(&static)
	return enable, err
}

//ISetRouterHandler 设置路由列表
type ISetRouterHandler interface {
	SetRouters([]*conf.Router) error
}

func SetRouters(engine servers.IExecuter, cnf conf.IServerConf, set ISetRouterHandler, ext map[string]interface{}) (enable bool, err error) {
	var routers conf.Routers
	_, err = cnf.GetSubObject("router", &routers)
	if err == conf.ErrNoSetting {
		err = fmt.Errorf("路由:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}

	for _, router := range routers.Routers {
		router.Handler = middleware.ContextHandler(engine, router.Name, router.Engine, router.Service, router.Setting, ext)
	}
	err = set.SetRouters(routers.Routers)
	if err != nil {
		return false, err
	}
	return len(routers.Routers) > 0, nil
}

//---------------------------------------------------------------------------
//-------------------------------view---------------------------------------
//---------------------------------------------------------------------------

//ISetView 设置view
type ISetView interface {
	SetView(*conf.View) error
}

//SetView 设置view
func SetView(set ISetView, cnf conf.IServerConf) (enable bool, err error) {
	//设置jwt安全认证参数
	var view conf.View
	if _, err = cnf.GetSubObject("view", &view); err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = set.SetView(&view)
	return err == nil, err
}

//ISetCircuitBreaker 设置CircuitBreaker
type ISetCircuitBreaker interface {
	CloseCircuitBreaker() error
	SetCircuitBreaker(*conf.CircuitBreaker) error
}

//SetCircuitBreaker 设置熔断配置
func SetCircuitBreaker(set ISetCircuitBreaker, cnf conf.IServerConf) (enable bool, err error) {
	//设置CircuitBreaker
	var breaker conf.CircuitBreaker
	if _, err = cnf.GetSubObject("circuit", &breaker); err == conf.ErrNoSetting || breaker.Disable {
		return false, set.CloseCircuitBreaker()
	}
	if err != nil {
		return false, err
	}
	err = set.SetCircuitBreaker(&breaker)
	return err == nil && !breaker.Disable, err
}

//---------------------------------------------------------------------------
//-------------------------------header---------------------------------------
//---------------------------------------------------------------------------

//ISetHeaderHandler 设置header
type ISetHeaderHandler interface {
	SetHeader(conf.Headers) error
}

//SetHeaders 设置header
func SetHeaders(set ISetHeaderHandler, cnf conf.IServerConf) (enable bool, err error) {
	//设置通用头信息
	var header conf.Headers
	if _, err = cnf.GetSubObject("header", &header); err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		err = fmt.Errorf("header配置有误:%v", err)
		return false, err
	}
	return len(header) > 0, set.SetHeader(header)
}

//---------------------------------------------------------------------------
//-------------------------------ajax---------------------------------------
//---------------------------------------------------------------------------

//IAjaxRequest 设置ajax
type IAjaxRequest interface {
	SetAjaxRequest(bool) error
}

//SetAjaxRequest 设置ajax
func SetAjaxRequest(set IAjaxRequest, cnf conf.IServerConf) (enable bool, err error) {
	if enable, err = cnf.GetBool("onlyAllowAjaxRequest", false); err != nil {
		return false, err
	}
	return enable, set.SetAjaxRequest(enable)
}

//---------------------------------------------------------------------------
//-------------------------------host---------------------------------------
//---------------------------------------------------------------------------

//ISetHosts 设置hosts
type ISetHosts interface {
	SetHosts(conf.Hosts) error
}

//SetHosts 设置hosts
func SetHosts(set ISetHosts, cnf conf.IServerConf) (enable bool, err error) {
	var hosts conf.Hosts
	hosts = cnf.GetStrings("host")
	return len(hosts) > 0, set.SetHosts(hosts)
}

//---------------------------------------------------------------------------
//-------------------------------jwt---------------------------------------
//---------------------------------------------------------------------------

//ISetJwtAuth 设置jwt
type ISetJwtAuth interface {
	SetJWT(*conf.Auth) error
}

//SetJWT 设置jwt
func SetJWT(set ISetJwtAuth, cnf conf.IServerConf) (enable bool, err error) {
	//设置jwt安全认证参数
	var auths conf.Authes
	var jwt *conf.Auth
	if _, err := cnf.GetSubObject("auth", &auths); err != nil && err != conf.ErrNoSetting {
		err = fmt.Errorf("jwt配置有误:%v", err)
		return false, err
	}
	if jwt, enable = auths["jwt"]; !enable {
		jwt = &conf.Auth{Disable: true}
	}
	err = set.SetJWT(jwt)
	return err == nil && !jwt.Disable, err
}