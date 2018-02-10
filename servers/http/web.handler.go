package http

import (
	x "net/http"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http/middleware"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

func (s *WebServer) getHandler(routers []*conf.Router) (h x.Handler, err error) {
	if !servers.IsDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	if s.views, err = s.loadHTMLGlob(engine); err != nil {
		s.Logger.Warnf("%s未找到模板:%v", s.conf.GetFullName(), err)
		return nil, err
	}
	engine.Use(middleware.Logging(s.conf)) //记录请求日志
	engine.Use(gin.Recovery())
	engine.Use(s.option.metric.Handle())           //生成metric报表
	engine.Use(middleware.Host(s.conf))            // 检查主机头是否合法
	engine.Use(middleware.Static(s.option.static)) //处理静态文件
	engine.Use(middleware.JwtAuth(s.conf))         //jwt安全认证
	engine.Use(middleware.Body())                  //处理请求form
	engine.Use(middleware.WebResponse(s.conf))     //处理返回值
	engine.Use(middleware.Header(s.conf))          //设置请求头
	if err = setRouters(engine, routers); err != nil {
		return nil, err
	}
	return engine, nil
}
func (s *WebServer) loadHTMLGlob(engine *gin.Engine) (files []string, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = err1.(error)
		}
	}()
	files = make([]string, 0, 8)
	viewPath := "./"
	if view, ok := s.conf.GetMeta("view").(*conf.View); ok {
		viewPath = view.Path
	}

	dirs := []string{path.Join(viewPath, "/*.html"), path.Join(viewPath, "/**/*.html")}
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
