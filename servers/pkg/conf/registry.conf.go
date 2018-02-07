package conf

import (
	"errors"
	"fmt"
	"strings"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/net"
)

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

var ERR_NOT_SETTING = errors.New("未配置")
var ERR_NO_CHANGED = errors.New("配置未变化")

type RegistryConf struct {
	nconf        xconf.Conf
	oconf        xconf.Conf
	Domain       string
	Name         string
	Type         string
	Cluster      string
	RegistryNode string
	IP           string
}

//GetFullName 获取服务名称
func (s *RegistryConf) GetFullName() string {
	return fmt.Sprintf("%s.%s(%s)", s.Name, s.Type, s.Cluster)
}
func NewRegistryConf(oconf xconf.Conf, nconf xconf.Conf) *RegistryConf {
	return &RegistryConf{
		nconf:        nconf,
		oconf:        oconf,
		Domain:       nconf.String("domain"),
		Name:         nconf.String("name"),
		Type:         nconf.String("type"),
		Cluster:      nconf.String("tag"),
		RegistryNode: nconf.Translate("{@category_path}/servers/{@tag}"),
		IP:           net.GetLocalIPAddress(nconf.String("mask")),
	}
}

//IsChanged 配置是否已经发生变化
func (s *RegistryConf) IsChanged() bool {
	return s.oconf.GetVersion() != s.nconf.GetVersion()
}

//IsStoped 服务器已停止
func (s *RegistryConf) IsStoped() bool {
	return strings.EqualFold(s.nconf.String("status"), servers.ST_STOP)
}

//GetHost 获取主机头
func (s *RegistryConf) GetHost() string {
	return s.nconf.String("host")
}

//GetMetric 获取metric配置
func (s *RegistryConf) GetMetric() (enable bool, host string, dataBase string, userName string, password string, cron string, err error) {
	//设置metric服务器监控数据
	metric, err := s.nconf.GetNodeWithSectionName("metric", "#@path/metric")
	if err != nil {
		if !s.nconf.Has("#@path/metric") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("metric未配置或配置有误:%+v", err)
		return false, "", "", "", "", "", err
	}
	enable = true
	if r, err := s.oconf.GetNodeWithSectionName("metric", "#@path/metric"); err != nil || r.GetVersion() != metric.GetVersion() {
		host := metric.String("host")
		dataBase := metric.String("dataBase")
		userName := metric.String("userName")
		password := metric.String("password")
		cron := metric.String("cron", "@every 1m")
		enable, _ = metric.Bool("enable", true)
		if host == "" || dataBase == "" {
			err = fmt.Errorf("metric配置错误:host 和 dataBase不能为空(`host:%s，dataBase:%s)", host, dataBase)
			return false, "", "", "", "", "", err
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		return enable, host, dataBase, userName, password, cron, nil
	}
	err = ERR_NO_CHANGED
	return

}

//GetAuth 获取安全认证配置参数
func (s *RegistryConf) GetAuth(name string) (a *Auth, err error) {
	a = &Auth{}
	auth, err := s.nconf.GetNodeWithSectionName("auth", "#@path/auth")
	if err != nil {
		if !s.nconf.Has("#@path/auth") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("auth配置有误:%+v", err)
		return a, err
	}
	if r, err := s.oconf.GetNodeWithSectionName("auth", "#@path/auth"); err != nil || r.GetVersion() != auth.GetVersion() {
		if !auth.Has(name) {
			err = ERR_NOT_SETTING
			return a, err
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
		return &Auth{Name: nm, Mode: mode, Secret: secret, Exclude: exclude, ExpireAt: int64(expireAt), Enable: enable}, nil
	}
	err = ERR_NO_CHANGED
	return
}

//GetOnlyAllowAjaxRequest 获取是否只允许ajax调用
func (s *RegistryConf) GetOnlyAllowAjaxRequest() bool {
	return s.nconf.String("onlyAllowAjaxRequest", "false") == "true"
}

//GetHeaders 获取http头信息
func (s *RegistryConf) GetHeaders() (hmap map[string]string, err error) {
	hmap = make(map[string]string)
	header, err := s.nconf.GetNodeWithSectionName("header", "#@path/header")
	if err != nil {
		if !s.nconf.Has("#@path/header") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("header未配置或配置有误:%+v", err)
		return nil, err
	}
	if r, err := s.oconf.GetNodeWithSectionName("header", "#@path/header"); err != nil || r.GetVersion() != header.GetVersion() {
		enable, _ := header.Bool("enable", true)
		if !enable {
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
	return hmap, ERR_NO_CHANGED
}
func (s *RegistryConf) GetRouters() (rrts []*Router, err error) {
	defAction := "get"
	routers, err := s.nconf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Router, 0, 4)
	if r, err := s.oconf.GetNodeWithSectionName("router", "#@path/router"); err != nil || r.GetVersion() != routers.GetVersion() {
		baseArgs := routers.String("args")
		rts, err := routers.GetSections("routers")
		if err != nil {
			return nil, fmt.Errorf("路由配置出错:err:%+v", err)
		}
		if len(rts) == 0 {
			return nil, ERR_NOT_SETTING
		}
		for _, c := range rts {
			name := c.String("name")
			service := c.String("service")
			actions := strings.Split(strings.ToUpper(c.String("action", defAction)), ",")
			engine := c.String("engine", "*")
			args := c.String("args")
			if name == "" || service == "" {
				return nil, fmt.Errorf("service 和 name不能为空（name:%s，service:%s）", name, service)
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
					return nil, fmt.Errorf("action:%v不支持,只支持:%v", actions, supportMethods)
				}
			}
			sigleRouter := &Router{
				Name:    name,
				Action:  actions,
				Engine:  engine,
				Service: service,
				Setting: baseArgs + "&" + args,
			}
			rrts = append(rrts, sigleRouter)
		}
		if len(rrts) == 0 {
			return nil, ERR_NOT_SETTING
		}
		return rrts, nil
	}
	return nil, ERR_NO_CHANGED
}

func (s *RegistryConf) GetServerRaw() (sr string, err error) {
	server, err := s.nconf.GetNodeWithSectionName("server", "#@path/server")
	if err != nil {
		return "", err
	}
	return server.GetContent(), nil
}

func (s *RegistryConf) GetQueues() (rrts []*Queue, err error) {
	routers, err := s.nconf.GetNodeWithSectionName("queue", "#@path/queue")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Queue, 0, 4)
	if r, err := s.oconf.GetNodeWithSectionName("queue", "#@path/queue"); err != nil || r.GetVersion() != routers.GetVersion() {
		baseArgs := routers.String("args")
		rts, err := routers.GetSections("queues")
		if err != nil {
			return nil, fmt.Errorf("queue配置出错:err:%+v", err)
		}
		if len(rts) == 0 {
			return nil, ERR_NOT_SETTING
		}
		for _, c := range rts {
			name := c.String("name")
			queue := c.String("queue", name)
			service := c.String("service")
			engine := c.String("engine", "*")
			concurrency, _ := c.Int("concurrency", 0)
			args := c.String("args")
			if name == "" || service == "" {
				return nil, fmt.Errorf("name 和 service不能为空（name:%s，service:%s）", name, service)
			}
			sigleRouter := &Queue{
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
	return nil, ERR_NO_CHANGED
}

func (s *RegistryConf) GetTasks() (rrts []*Task, err error) {
	tasks, err := s.nconf.GetNodeWithSectionName("task", "#@path/task")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Task, 0, 4)
	if r, err := s.oconf.GetNodeWithSectionName("task", "#@path/task"); err != nil || r.GetVersion() != tasks.GetVersion() {
		baseArgs := tasks.String("args")
		rts, err := tasks.GetSections("tasks")
		if err != nil {
			return nil, fmt.Errorf("task配置出错:err:%+v", err)
		}
		if len(rts) == 0 {
			return nil, ERR_NOT_SETTING
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
				return nil, fmt.Errorf("name,cron,service不能为空（name:%s，cron:%s,service:%s）", name, cron, service)
			}
			sigleRouter := &Task{
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
	return nil, ERR_NO_CHANGED
}
func (s *RegistryConf) GetRedisRaw() (sr string, err error) {
	redis, err := s.nconf.GetNodeWithSectionName("redis", "#@path/redis")
	if err != nil {
		return "", err
	}
	return redis.GetContent(), nil
}
