package standard

import (
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//SetRouters 设置路由配置
func (s *Server) SetRouters(routers []*conf.Router) (err error) {
	s.Processor, err = s.getProcessor(routers)
	if err != nil {
		return
	}
	return
}
