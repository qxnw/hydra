package standard

import (
	"fmt"

	"github.com/qxnw/hydra/servers/pkg/conf"
	"github.com/qxnw/hydra/servers/pkg/middleware"
)

func (s *Server) getProcessor(addr string, raw string, queues []*conf.Queue) (engine *Processor, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("%v", err1)
		}
	}()
	engine, err = NewProcessor(addr, raw, queues)
	if err != nil {
		return nil, err
	}
	engine.Use(middleware.Logging(s.conf)) //记录请求日志
	engine.Use(middleware.Recovery())
	engine.Use(s.option.metric.Handle()) //生成metric报表
	engine.Use(middleware.NoResponse(s.conf))
	engine.AddRouters()
	err = engine.Consumes()
	if err != nil {
		return nil, err
	}
	return engine, nil
}
