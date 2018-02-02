package standard

import (
	"fmt"
	x "net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

func (s *Server) getHandler(routers []*conf.Router) (x.Handler, error) {
	if !servers.IsDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	views, err := s.loadHTMLGlob(engine)
	if err != nil {
		s.Logger.Warnf("%s未找到模板:%v", s.conf.GetFullName(), err)
		return nil, err
	}
	s.views = views
	engine.Use(middleware.Logging(s.conf.ServerConf)) //记录请求日志
	engine.Use(gin.Recovery())
	engine.Use(s.option.metric.Handle())                  //生成metric报表
	engine.Use(middleware.Host(s.conf.ServerConf))        // 检查主机头是否合法
	engine.Use(middleware.Static(s.option.static))        //处理静态文件
	engine.Use(middleware.JwtAuth(s.conf.ServerConf))     //jwt安全认证
	engine.Use(middleware.Body())                         //处理请求form
	engine.Use(middleware.WebResponse(s.conf.ServerConf)) //处理返回值
	engine.Use(middleware.Header(s.conf.ServerConf))      //设置请求头
	err = setRouters(engine, routers)
	if err != nil {
		return nil, err
	}
	return engine, nil
}
func (s *Server) loadHTMLGlob(engine *gin.Engine) (files []string, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = err1.(error)
		}
	}()
	files = make([]string, 0, 8)
	dirs := []string{path.Join(s.conf.View.Path, "/*.html"), path.Join(s.conf.View.Path, "/**/*.html")}
	for _, name := range dirs {
		filenames, err := filepath.Glob(name)
		if err != nil {
			return nil, err
		}
		files = append(files, filenames...)
	}
	engine.LoadHTMLFiles(files...)
	return nil, nil
}

func setRouters(engine *gin.Engine, routers []*conf.Router) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("%v", err1)
		}
	}()
	for _, router := range routers {
		for _, method := range router.Action {
			engine.Handle(strings.ToUpper(method), router.Name, router.Handler.(gin.HandlerFunc))
		}
	}
	return err
}

type Routers struct {
	routers []*conf.Router
}

func GetRouters() *Routers {
	return &Routers{
		routers: make([]*conf.Router, 0, 2),
	}
}
func (r *Routers) Get() []*conf.Router {
	return r.routers
}
func (r *Routers) Route(method string, name string, f servers.IExecuteHandler) {
	r.routers = append(r.routers,
		&conf.Router{
			Name:    name,
			Action:  strings.Split(method, ","),
			Engine:  "*",
			Service: name,
			Handler: middleware.ContextHandler(f, name, "*", name, ""),
		})
}
