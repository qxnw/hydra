package standard

import (
	x "net/http"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/http"
	"github.com/qxnw/hydra/servers/http/middleware"
)

func (s *Server) getHandler(routers []*http.Router) x.Handler {
	s.Debugf("routers:",len(routers))
	if servers.IsDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logging(s.conf))
	engine.Use(s.option.metric.Handle())
	engine.Use(middleware.Host(s.conf))
	engine.Use(middleware.Header(s.conf))
	engine.Use(middleware.Static(s.option.static))
	engine.Use(middleware.AjaxRequest(s.conf))
	engine.Use(middleware.JwtAuth(s.conf))
	engine.Use(middleware.Body())
	engine.Use(middleware.APIResponse(s.conf))
	setRouters(engine, routers)
	return engine
}
func setRouters(engine *gin.Engine, routers []*http.Router) {
	for _, router := range routers {
		for _, method := range router.Action {
			engine.Handle(method, router.Name, router.Handler.(gin.HandlerFunc))
		}
	}
}
