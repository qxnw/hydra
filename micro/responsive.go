package micro

import (
	"sync"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry/watcher"
	"github.com/qxnw/lib4go/logger"
)

type responsiveServers struct {
	servers      map[string]*server
	mu           sync.Mutex
	registry     registry.IRegistry
	registryAddr string
	logger       *logger.Logger
	done         bool
}

func newResponsiveServers(registryAddr string, registry registry.IRegistry, logger *logger.Logger) *responsiveServers {
	return &responsiveServers{
		registry:     registry,
		registryAddr: registryAddr,
		servers:      make(map[string]*server),
		logger:       logger,
	}
}

//Change 服务器发生变更
func (s *responsiveServers) Change(u *watcher.ContentChangeArgs) {
	if s.done {
		return
	}
	switch u.OP {
	case watcher.ADD, watcher.CHANGE:
		s.mu.Lock()
		defer s.mu.Unlock()
		//初始化服务器配置
		conf, err := conf.NewServerConf(u.Path, u.Content, u.Version, s.registry)
		if err != nil {
			s.logger.Errorf("%s配置有误:%v", u.Path, err)
			return
		}
		if _, ok := s.servers[u.Path]; !ok {
			//添加新服务器
			server := newServer(conf, s.registryAddr, s.registry)
			if err = server.Start(); err != nil {
				s.logger.Errorf("%s启动失败:%v", conf.GetSysName(), err)
				return
			}
			s.servers[u.Path] = server
		} else {
			//修改服务器
			server := s.servers[u.Path]
			if err = server.Notify(conf); err != nil {
				server.Shutdown()
				delete(s.servers, u.Path)
				s.logger.Errorf("%s配置更新失败:%v", conf.GetSysName(), err)
				return
			}
		}

	case watcher.DEL:
		s.mu.Lock()
		defer s.mu.Unlock()
		if server, ok := s.servers[u.Path]; ok {
			s.logger.Errorf("%s配置已删除", u.Path)
			server.Shutdown()
			delete(s.servers, u.Path)
			return
		}
	}
}

//Change 服务器发生变更
func (s *responsiveServers) Shutdown() {
	s.done = true
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, server := range s.servers {
		server.Shutdown()
	}
}
