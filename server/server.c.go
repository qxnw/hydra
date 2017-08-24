package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/conf"
)

var ERR_NOT_SETTING = errors.New("未配置")
var ERR_NO_CHANGED = errors.New("配置未变化")

//GetMetric 获取metric配置
func GetMetric(oconf conf.Conf, nconf conf.Conf) (enable bool, host string, dataBase string, userName string, password string, span time.Duration, err error) {
	//设置metric服务器监控数据
	metric, err := nconf.GetNodeWithSectionName("metric", "#@path/metric")
	if err != nil {
		if !nconf.Has("#@path/metric") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("metric未配置或配置有误:%+v", err)
		return false, "", "", "", "", 0, err
	}
	enable = true
	if r, err := oconf.GetNodeWithSectionName("metric", "#@path/metric"); err != nil || r.GetVersion() != metric.GetVersion() {
		host := metric.String("host")
		dataBase := metric.String("dataBase")
		userName := metric.String("userName")
		password := metric.String("password")
		enable, _ = metric.Bool("enable", true)
		if host == "" || dataBase == "" {
			err = fmt.Errorf("metric配置错误:host 和 dataBase不能为空(`host:%s，dataBase:%s)", host, dataBase)
			return false, "", "", "", "", 0, err
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		return enable, host, dataBase, userName, password, time.Second * 60, nil
	}
	err = ERR_NO_CHANGED
	return

}

//GetStatic 获取静态文件配置内容
func GetStatic(oconf conf.Conf, nconf conf.Conf) (enable bool, prefix, dir string, showDir bool, exts []string, err error) {

	static, err := nconf.GetNodeWithSectionName("static", "#@path/static")
	if err != nil {
		if !nconf.Has("#@path/static") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("static未配置或配置有误:%+v", err)
		return false, "", "", false, nil, err
	}
	enable = true
	if r, err := oconf.GetNodeWithSectionName("static", "#@path/static"); err != nil || r.GetVersion() != static.GetVersion() {
		prefix := static.String("prefix")
		dir := static.String("dir")
		showDir := static.String("showDir") == "true"
		exts := static.Strings("exts")
		enable, _ = static.Bool("enable", true)
		if dir == "" {
			err = errors.New("static配置错误：dir不能为空")
			return false, prefix, dir, showDir, exts, err
		}
		return enable, prefix, dir, showDir, exts, nil
	}
	err = ERR_NO_CHANGED
	return
}

//GetXSRF 获取静态文件配置内容
func GetXSRF(oconf conf.Conf, nconf conf.Conf) (enable bool, key, secret string, err error) {
	xsrf, err := nconf.GetNodeWithSectionName("xsrf", "#@path/xsrf")
	if err != nil {
		if !nconf.Has("#@path/xsrf") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("xsrf未配置或配置有误:%+v", err)
		return false, "", "", err
	}
	enable = true
	if r, err := oconf.GetNodeWithSectionName("xsrf", "#@path/xsrf"); err != nil || r.GetVersion() != xsrf.GetVersion() {
		key := xsrf.String("key")
		secret := xsrf.String("secret")
		enable, _ := xsrf.Bool("enable", true)
		if key == "" || secret == "" {
			err = fmt.Errorf("xsrf配置错误：key,secret不能为空(%s,%s,%s)", xsrf.String("name"), key, secret)
			return false, "", "", err
		}
		return enable, key, secret, nil
	}
	err = ERR_NO_CHANGED
	return
}

//GetOnlyAllowAjaxRequest 获取是否只允许ajax调用
func GetOnlyAllowAjaxRequest(nconf conf.Conf) bool {
	return nconf.String("onlyAllowAjaxRequest", "false") == "true"
}

//GetHeaders 获取http头信息
func GetHeaders(oconf conf.Conf, nconf conf.Conf) (hmap map[string]string, err error) {
	hmap = make(map[string]string)
	header, err := nconf.GetNodeWithSectionName("header", "#@path/header")
	if err != nil {
		if !nconf.Has("#@path/header") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("header未配置或配置有误:%+v", err)
		return nil, err
	}
	if r, err := oconf.GetNodeWithSectionName("header", "#@path/header"); err != nil || r.GetVersion() != header.GetVersion() {
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
func GetRouters(oconf conf.Conf, nconf conf.Conf, defAction string, supportMethods []string) (rrts []*Router, err error) {
	routers, err := nconf.GetNodeWithSectionName("router", "#@path/router")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Router, 0, 4)
	if r, err := oconf.GetNodeWithSectionName("router", "#@path/router"); err != nil || r.GetVersion() != routers.GetVersion() {
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
			mode := c.String("mode", "go")
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
				Mode:    mode,
				Service: service,
				Args:    baseArgs + "&" + args,
			}
			rrts = append(rrts, sigleRouter)
		}
		if len(rrts) == 0 {
			return nil, fmt.Errorf("路由未配置:%d", len(rrts))
		}
		return rrts, nil
	}
	return nil, ERR_NO_CHANGED
}

func GetTasks(oconf conf.Conf, nconf conf.Conf) (rrts []*Task, err error) {
	tasks, err := nconf.GetNodeWithSectionName("task", "#@path/task")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Task, 0, 4)
	if r, err := oconf.GetNodeWithSectionName("task", "#@path/task"); err != nil || r.GetVersion() != tasks.GetVersion() {
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
			mode := c.String("mode", "go")
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
				Mode:    mode,
				Input:   input,
				Body:    body,
				Service: service,
				Args:    baseArgs + "&" + args,
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

func GetQueues(oconf conf.Conf, nconf conf.Conf) (rrts []*Queue, err error) {
	routers, err := nconf.GetNodeWithSectionName("queue", "#@path/queue")
	if err != nil {
		return nil, err
	}
	rrts = make([]*Queue, 0, 4)
	if r, err := oconf.GetNodeWithSectionName("queue", "#@path/queue"); err != nil || r.GetVersion() != routers.GetVersion() {
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
			service := c.String("service")
			mode := c.String("mode", "go")
			args := c.String("args")
			if name == "" || service == "" {
				return nil, fmt.Errorf("name 和 service不能为空（name:%s，service:%s）", name, service)
			}
			sigleRouter := &Queue{
				Name:    name,
				Mode:    mode,
				Service: service,
				Args:    baseArgs + "&" + args,
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

//GetLimiters 获取限量规则
func GetLimiters(oconf conf.Conf, nconf conf.Conf) (rrts []*Limiter, err error) {
	routers, err := nconf.GetNodeWithSectionName("limiter", "#@path/limiter")
	if err != nil {
		if !nconf.Has("#@path/limiter") {
			err = ERR_NOT_SETTING
			return
		}
		return nil, err
	}
	rrts = make([]*Limiter, 0, 4)
	if r, err := oconf.GetNodeWithSectionName("limiter", "#@path/limiter"); err != nil || r.GetVersion() != routers.GetVersion() {
		rts, err := routers.GetSections("limiters")
		if err != nil {
			return nil, fmt.Errorf("limiter配置出错:err:%+v", err)
		}
		if len(rts) == 0 {
			return nil, ERR_NOT_SETTING
		}
		for _, c := range rts {
			name := c.String("name")
			value, err := c.Int("value")
			if err != nil || name == "" || value == 0 {
				return nil, fmt.Errorf("name 和 value不能为空（name:%s，value:%d）(err:%v)", name, value, err)
			}
			sigleRouter := &Limiter{
				Name:  name,
				Value: value,
			}
			rrts = append(rrts, sigleRouter)
		}
		if len(rrts) == 0 {
			return nil, fmt.Errorf("limiter未配置:%d", len(rrts))
		}
		return rrts, nil
	}
	return nil, ERR_NO_CHANGED
}

//GetViews 获取View配置
func GetViews(oconf conf.Conf, nconf conf.Conf) (rrts *View, err error) {
	xsrf, err := nconf.GetNodeWithSectionName("view", "#@path/view")
	if err != nil {
		if !nconf.Has("#@path/view") {
			err = ERR_NOT_SETTING
			return
		}
		err = fmt.Errorf("view未配置或配置有误:%+v", err)
		return nil, err
	}
	if r, err := oconf.GetNodeWithSectionName("xsrf", "#@path/xsrf"); err != nil || r.GetVersion() != xsrf.GetVersion() {
		viewPath := xsrf.String("viewPath", "../views")
		left := xsrf.String("left", "{{")
		right := xsrf.String("right", "}}")
		return &View{ViewPath: viewPath, Left: left, Right: right}, nil
	}
	err = ERR_NO_CHANGED
	return
}

type Router struct {
	Name    string
	Action  []string
	Mode    string
	Service string
	Args    string
}
type Task struct {
	Name    string
	Cron    string
	Input   string
	Body    string
	Mode    string
	Service string
	Args    string
}
type Queue struct {
	Name    string
	Mode    string
	Service string
	Args    string
}

type Limiter struct {
	Name  string
	Value int
}
type View struct {
	ViewPath string
	Left     string
	Right    string
}

/*
	path := view.String("viewPath", "../views")
			left := view.String("left", "{{")
			right := view.String("right", "}}")
*/
