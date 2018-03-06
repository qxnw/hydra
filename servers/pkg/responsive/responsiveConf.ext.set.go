package responsive

import (
	"fmt"

	"github.com/qxnw/hydra/servers"
	hmiddleware "github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
)

//---------------------------------------------------------------------------
//-------------------------------header---------------------------------------
//---------------------------------------------------------------------------

//ISetHeaderHandler 设置header
type ISetHeaderHandler interface {
	SetHeader(map[string]string) error
}

//SetHeaders 设置header
func (s *ResponsiveConf) SetHeaders(set ISetHeaderHandler) (enable bool, err error) {
	//设置通用头信息
	headers, err := s.GetHeaders()
	if err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		err = fmt.Errorf("header配置有误:%v", err)
		return false, err
	}
	return len(headers) > 0, set.SetHeader(headers)
}

//---------------------------------------------------------------------------
//-------------------------------ajax---------------------------------------
//---------------------------------------------------------------------------

//IAjaxRequest 设置ajax
type IAjaxRequest interface {
	SetAjaxRequest(bool) error
}

//SetAjaxRequest 设置ajax
func (s *ResponsiveConf) SetAjaxRequest(set IAjaxRequest) (enable bool, err error) {
	enable = s.GetOnlyAllowAjaxRequest()
	return enable, set.SetAjaxRequest(enable)
}

//---------------------------------------------------------------------------
//-------------------------------host---------------------------------------
//---------------------------------------------------------------------------

//ISetHosts 设置hosts
type ISetHosts interface {
	SetHosts([]string) error
}

//SetHosts 设置hosts
func (s *ResponsiveConf) SetHosts(set ISetHosts) (enable bool, err error) {
	hosts := s.GetHosts()
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
func (s *ResponsiveConf) SetJWT(set ISetJwtAuth) (enable bool, err error) {
	//设置jwt安全认证参数
	jwt, err := s.GetAuth("jwt")
	if err == conf.ErrNoSetting {
		err = nil
		jwt = &conf.Auth{Enable: false}
	}
	if err != nil {
		err = fmt.Errorf("jwt配置有误:%v", err)
		return false, err
	}
	err = set.SetJWT(jwt)
	return err == nil && jwt.Enable, err
}

//---------------------------------------------------------------------------
//-------------------------------router---------------------------------------
//---------------------------------------------------------------------------

var supportMethods = []string{
	"GET",
	"POST",
	"HEAD",
	"DELETE",
	"PUT",
	"OPTIONS",
	"TRACE",
	"PATCH",
}

//ISetRouterHandler 设置路由列表
type ISetRouterHandler interface {
	SetRouters([]*conf.Router) error
}

//SetHttpRouters 设置路由
func (s *ResponsiveConf) SetHttpRouters(engine servers.IExecuter, set ISetRouterHandler, ext map[string]interface{}) (enable bool, err error) {
	routers, err := s.GetRouters()
	if err == conf.ErrNoSetting {
		err = fmt.Errorf("路由:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}

	for _, router := range routers {
		router.Handler = hmiddleware.ContextHandler(engine, router.Name, router.Engine, router.Service, router.Setting)
	}
	err = set.SetRouters(routers)
	if err != nil {
		return false, err
	}
	return len(routers) > 0, nil
}

//SetRouters 设置路由
func (s *ResponsiveConf) SetRouters(engine servers.IExecuter, set ISetRouterHandler, ext map[string]interface{}) (enable bool, err error) {
	routers, err := s.GetRouters()
	if err == conf.ErrNoSetting {
		err = fmt.Errorf("路由:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}

	for _, router := range routers {
		router.Handler = middleware.ContextHandler(engine, router.Name, router.Engine, router.Service, router.Setting, nil)
	}
	err = set.SetRouters(routers)
	if err != nil {
		return false, err
	}
	return len(routers) > 0, nil
}

//---------------------------------------------------------------------------
//-------------------------------static---------------------------------------
//---------------------------------------------------------------------------

//ISetStatic 设置static
type ISetStatic interface {
	SetStatic(enable bool, prefix string, dir string, listDir bool, exts []string) error
}

//SetStatic 设置static
func (s *ResponsiveConf) SetStatic(set ISetStatic) (enable bool, err error) {
	//设置静态文件路由
	enable, prefix, dir, showDir, exts, err := s.GetStatic()
	if err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = set.SetStatic(enable, prefix, dir, showDir, exts)
	return enable, err
}

//---------------------------------------------------------------------------
//-------------------------------view---------------------------------------
//---------------------------------------------------------------------------

//ISetView 设置view
type ISetView interface {
	SetView(*conf.View) error
}

//SetView 设置view
func (s *ResponsiveConf) SetView(set ISetView) (enable bool, err error) {
	//设置jwt安全认证参数
	view, err := s.GetView()
	if err == conf.ErrNoSetting {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = set.SetView(view)
	return err == nil, err
}

//ITasks 设置tasks
type ITasks interface {
	SetTasks(string, []*conf.Task) error
}

//SetTasks 设置tasks
func (s *ResponsiveConf) SetTasks(engine servers.IExecuter, set ITasks, ext map[string]interface{}) (enable bool, err error) {

	redisConf, _ := s.GetRedisRaw()
	tasks, err := s.GetTasks()
	if err == conf.ErrNoSetting || len(tasks) == 0 {
		err = conf.ErrNoSetting
		err = fmt.Errorf("task:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}
	if ext == nil {
		ext = map[string]interface{}{}
	}
	for _, task := range tasks {
		ext["__cron_"] = task.Cron
		task.Handler = middleware.ContextHandler(engine, task.Name, task.Engine, task.Service, task.Setting, ext)
	}
	err = set.SetTasks(redisConf, tasks)
	if err != nil {
		return false, err
	}
	return true, nil
}

//IQueues 设置queue
type IQueues interface {
	SetQueues(string, []*conf.Queue) error
}

//SetQueues 设置queue
func (s *ResponsiveConf) SetQueues(engine servers.IExecuter, set IQueues, ext map[string]interface{}) (enable bool, err error) {

	//设置queue
	queues, err := s.GetQueues()
	if err == conf.ErrNoSetting || len(queues) == 0 {
		err = conf.ErrNoSetting
		err = fmt.Errorf("queue:%v", err)
		return false, err
	}
	if err != nil {
		return false, err
	}
	if ext == nil {
		ext = map[string]interface{}{}
	}
	for _, queue := range queues {
		queue.Handler = middleware.ContextHandler(engine, queue.Name, queue.Engine, queue.Service, queue.Setting, ext)
	}
	serverRaw, err := s.GetServerRaw()
	if err == conf.ErrNoSetting {
		err = fmt.Errorf("server节点:%v", err)
		return false, err
	}
	if err != nil {
		return
	}
	err = set.SetQueues(serverRaw, queues)
	if err != nil {
		return false, err
	}
	return true, nil
}

//ISetCircuitBreaker 设置CircuitBreaker
type ISetCircuitBreaker interface {
	CloseCircuitBreaker() error
	SetCircuitBreaker(*conf.CircuitBreaker) error
}

//SetCircuitBreaker 设置熔断配置
func (s *ResponsiveConf) SetCircuitBreaker(set ISetCircuitBreaker) (enable bool, err error) {
	//设置CircuitBreaker
	breaker, err := s.GetCircuitBreaker()
	if err == conf.ErrNoSetting || !breaker.Enable {
		return false, set.CloseCircuitBreaker()
	}
	if err != nil {
		return false, err
	}
	err = set.SetCircuitBreaker(breaker)
	return err == nil && breaker.Enable, err
}
