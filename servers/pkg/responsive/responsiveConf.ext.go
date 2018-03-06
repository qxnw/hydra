package responsive

import (
	"errors"
	"fmt"
	"strings"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//ResponsiveConf api响应式配置
type ResponsiveConf struct {
	*conf.ServerConf
	*ResponsiveBaseConf
	OnlyAllowAjaxRequest bool
}

//NewResponsiveConfBy 创建api 响应式配置
func NewResponsiveConfBy(oconf xconf.Conf, nconf xconf.Conf) *ResponsiveConf {
	return &ResponsiveConf{
		ServerConf:         conf.NewConfBy(nconf),
		ResponsiveBaseConf: NewResponsiveBaseConf(oconf, nconf),
	}
}

//CopyNew 根据当前conf复制一个新的conf
func (s *ResponsiveConf) CopyNew(nconf xconf.Conf) *ResponsiveConf {
	return &ResponsiveConf{
		ServerConf:         s.ServerConf,
		ResponsiveBaseConf: NewResponsiveBaseConf(s.Nconf, nconf),
	}
}

//GetHosts 获取主机头
func (s *ResponsiveConf) GetHosts() []string {
	host := s.Nconf.String("host")
	if host == "" {
		return nil
	}
	return strings.Split(host, ";")
}

//GetAuth 获取安全认证配置参数
func (s *ResponsiveConf) GetAuth(name string) (a *conf.Auth, err error) {
	a = &conf.Auth{}
	auth, err := s.Nconf.GetNodeWithSectionName("auth", "#@path/auth")
	if err != nil {
		if !s.Nconf.Has("#@path/auth") {
			return nil, conf.ErrNoSetting
		}
		err = fmt.Errorf("auth[%s]配置有误:%+v", name, err)
		return nil, err
	}
	if !auth.Has(name) {
		return nil, conf.ErrNoSetting
	}
	xsrf, err := auth.GetSection(name)
	if err != nil {
		return a, err
	}
	nm := xsrf.String("name")
	mode := xsrf.String("mode", "HS512")
	secret := xsrf.String("secret")
	exclude := xsrf.Strings("exclude")
	expireAt, _ := xsrf.Int("expireAt", 0)
	enable, _ := xsrf.Bool("enable", true)
	return &conf.Auth{Name: nm, Mode: mode, Secret: secret, Exclude: exclude, ExpireAt: int64(expireAt), Enable: enable}, nil

}

//GetOnlyAllowAjaxRequest 获取是否只允许ajax调用
func (s *ResponsiveConf) GetOnlyAllowAjaxRequest() bool {
	return s.Nconf.String("onlyAllowAjaxRequest", "false") == "true"
}

//GetHeaders 获取http头信息
func (s *ResponsiveConf) GetHeaders() (hmap map[string]string, err error) {
	hmap = make(map[string]string)
	header, err := s.Nconf.GetNodeWithSectionName("header", "#@path/header")
	if err != nil {
		if !s.Nconf.Has("#@path/header") {
			err = conf.ErrNoSetting
			return
		}
		err = fmt.Errorf("header未配置或配置有误:%+v", err)
		return nil, err
	}
	if enable, _ := header.Bool("enable", true); !enable {
		return hmap, nil
	}
	list := header.GetData()
	for k, v := range list {
		if k != "enable" {
			hmap[k] = fmt.Sprint(v)
		}
	}
	return hmap, nil

}

//GetRouters 获取路由配置
func (s *ResponsiveConf) GetRouters() (rrts []*conf.Router, err error) {
	if !s.HasNode("router") {
		return nil, conf.ErrNoSetting
	}
	defAction := "get"
	routers, err := s.Nconf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return nil, fmt.Errorf("路由配置有误:%v", err)
	}
	rrts = make([]*conf.Router, 0, 4)
	baseArgs := routers.String("args")
	rts, err := routers.GetSections("routers")
	if err != nil {
		return nil, err
	}
	if len(rts) == 0 {
		return nil, conf.ErrNoSetting
	}
	for _, c := range rts {
		name := c.String("name")
		service := c.String("service")
		actions := strings.Split(strings.ToUpper(c.String("action", defAction)), ",")
		engine := c.String("engine", "*")
		args := c.String("args")
		if name == "" || service == "" {
			return nil, fmt.Errorf("router配置出错:service 和 name不能为空（name:%s，service:%s）", name, service)
		}
		for _, v := range actions {
			exist := false
			for _, e := range supportMethods {
				if v == e {
					exist = true
					break
				}
			}
			if !exist {
				return nil, fmt.Errorf("router配置出错:action:%v不支持,只支持:%v", actions, supportMethods)
			}
		}
		sigleRouter := &conf.Router{
			Name:    name,
			Action:  actions,
			Engine:  engine,
			Service: service,
			Setting: baseArgs + "&" + args,
		}
		rrts = append(rrts, sigleRouter)
	}
	if len(rrts) == 0 {
		return nil, conf.ErrNoSetting
	}
	return rrts, nil
}

//GetStatic 获取静态文件配置内容
func (s *ResponsiveConf) GetStatic() (enable bool, prefix, dir string, showDir bool, exts []string, err error) {
	static, err := s.Nconf.GetNodeWithSectionName("static", "#@path/static")
	if err != nil {
		if !s.Nconf.Has("#@path/static") {
			err = conf.ErrNoSetting
			return
		}
		err = fmt.Errorf("static未配置或配置有误:%+v", err)
		return false, "", "", false, nil, err
	}
	enable = false
	prefix = static.String("prefix")
	dir = static.String("dir")
	showDir = static.String("showDir") == "true"
	exts = static.Strings("exts")
	enable, _ = static.Bool("enable", true)
	if dir == "" || prefix == "" {
		err = errors.New("static配置错误：dir,prefix不能为空")
		return false, prefix, dir, showDir, exts, err
	}
	return enable, prefix, dir, showDir, exts, nil

}

//GetView 获取View配置
func (s *ResponsiveConf) GetView() (rrts *conf.View, err error) {
	xsrf, err := s.Nconf.GetNodeWithSectionName("view", "#@path/view")
	if err != nil {
		if !s.Nconf.Has("#@path/view") {
			return nil, conf.ErrNoSetting
		}
		return nil, fmt.Errorf("view未配置或配置有误:%+v", err)
	}
	viewPath := xsrf.String("path", "../views")
	left := xsrf.String("left", "{{")
	right := xsrf.String("right", "}}")
	return &conf.View{Path: viewPath, Left: left, Right: right}, nil

}

//GetServerRaw 获取server节点配置
func (s *ResponsiveConf) GetServerRaw() (sr string, err error) {
	if !s.HasNode("server") {
		return "", conf.ErrNoSetting
	}
	server, err := s.Nconf.GetNodeWithSectionName("server", "#@path/server")
	if err != nil {
		err = fmt.Errorf("server配置有误:err:%+v", err)
		return "", err
	}
	return server.GetContent(), nil
}

//GetQueues 获取消息队列配置
func (s *ResponsiveConf) GetQueues() (rrts []*conf.Queue, err error) {
	routers, err := s.Nconf.GetNodeWithSectionName("queue", "#@path/queue")
	if err != nil {
		err = fmt.Errorf("queue配置出错:err:%+v", err)
		return nil, err
	}
	rrts = make([]*conf.Queue, 0, 4)

	baseArgs := routers.String("args")
	rts, err := routers.GetSections("queues")
	if err != nil {
		return nil, fmt.Errorf("queue配置出错:err:%+v", err)
	}
	if len(rts) == 0 {
		return nil, conf.ErrNoSetting
	}
	for _, c := range rts {
		name := c.String("name")
		queue := c.String("queue", name)
		service := c.String("service")
		engine := c.String("engine", "*")
		concurrency, _ := c.Int("concurrency", 0)
		args := c.String("args")
		if name == "" || service == "" {
			return nil, fmt.Errorf("queue配置出错:name 和 service不能为空（name:%s，service:%s）", name, service)
		}
		sigleRouter := &conf.Queue{
			Name:        name,
			Queue:       queue,
			Concurrency: concurrency,
			Engine:      engine,
			Service:     service,
			Setting:     baseArgs + "&" + args,
		}
		rrts = append(rrts, sigleRouter)
	}
	if len(rrts) == 0 {
		return nil, fmt.Errorf("queue未配置:%d", len(rrts))
	}
	return rrts, nil

}

//GetTasks 获取任务配置
func (s *ResponsiveConf) GetTasks() (rrts []*conf.Task, err error) {
	if !s.HasNode("task") {
		return nil, conf.ErrNoSetting
	}
	tasks, err := s.Nconf.GetNodeWithSectionName("task", "#@path/task")
	if err != nil {
		return nil, err
	}
	rrts = make([]*conf.Task, 0, 4)
	baseArgs := tasks.String("args")
	rts, err := tasks.GetSections("tasks")
	if err != nil {
		return nil, fmt.Errorf("task配置出错:err:%+v", err)
	}
	if len(rts) == 0 {
		return nil, conf.ErrNoSetting
	}
	for _, c := range rts {
		name := c.String("name")
		service := c.String("service")
		engine := c.String("engine", "*")
		args := c.String("args")
		input := c.String("input")
		body := c.String("body")
		cron := c.String("cron")
		if name == "" || service == "" || cron == "" {
			return nil, fmt.Errorf("task配置出错:task配置name,cron,service不能为空（name:%s，cron:%s,service:%s）", name, cron, service)
		}
		sigleRouter := &conf.Task{
			Name:    name,
			Cron:    cron,
			Engine:  engine,
			Input:   input,
			Body:    body,
			Service: service,
			Setting: baseArgs + "&" + args,
		}
		rrts = append(rrts, sigleRouter)
	}
	if len(rrts) == 0 {
		return nil, fmt.Errorf("task未配置:%d", len(rrts))
	}
	return rrts, nil

}

//GetRedisRaw 获取redis配置
func (s *ResponsiveConf) GetRedisRaw() (sr string, err error) {
	redis, err := s.Nconf.GetNodeWithSectionName("redis", "#@path/redis")
	if err != nil {
		return "", err
	}
	return redis.GetContent(), nil
}

//GetCircuitBreaker 熔断配置
func (s *ResponsiveConf) GetCircuitBreaker() (breaker *conf.CircuitBreaker, err error) {
	if !s.HasNode("circuit") {
		return nil, conf.ErrNoSetting
	}
	tasks, err := s.Nconf.GetNodeWithSectionName("circuit", "#@path/circuit")
	if err != nil {
		return nil, err
	}
	breaker = &conf.CircuitBreaker{}
	breaker.Enable, _ = tasks.Bool("enable", true)
	breaker.ForceBreak, _ = tasks.Bool("force-break", false)
	breaker.SwitchWindow, _ = tasks.Int("swith-window", 0)
	if !breaker.Enable {
		return breaker, nil
	}
	breakers, err := tasks.GetSections("circuit-breakers")
	if err != nil {
		return nil, fmt.Errorf("circuit-breakers配置出错:err:%+v", err)
	}
	if len(breakers) == 0 {
		return nil, conf.ErrNoSetting
	}
	breaker.CircuitBreakers = make(map[string]*conf.Breaker)
	for _, c := range breakers {
		oneBreaker := &conf.Breaker{}
		oneBreaker.URL = c.String("url")
		if _, ok := breaker.CircuitBreakers[oneBreaker.URL]; ok {
			return nil, fmt.Errorf("circuit-breakers重复配置:url:%s", c.String("url"))
		}

		oneBreaker.RequestPerSecond, _ = c.Int("request-per-second", 0)
		oneBreaker.RejectPerSecond, _ = c.Int("reject-per-second", 0)
		oneBreaker.FailedPercent, _ = c.Int("failed-percent", 0)
		if oneBreaker.RequestPerSecond < 0 || oneBreaker.RejectPerSecond < 0 || oneBreaker.FailedPercent < 0 {
			return nil, fmt.Errorf("circuit-breakers配置出错:%s", c.GetContent())
		}
		breaker.CircuitBreakers[oneBreaker.URL] = oneBreaker
	}
	if len(breakers) == 0 {
		return nil, fmt.Errorf("circuit-breakers未配置:%d", len(breakers))
	}
	return breaker, nil
}
