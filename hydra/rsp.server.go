package hydra

import (
	"sync"

	"github.com/qxnw/hydra/servers"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/registry/watcher"
	"github.com/qxnw/lib4go/logger"
)

type rspServer struct {
	servers      map[string]*server
	mu           sync.Mutex
	registry     registry.IRegistry
	registryAddr string
	logger       *logger.Logger
	done         bool
}

func newRspServer(registryAddr string, registry registry.IRegistry, logger *logger.Logger) *rspServer {
	return &rspServer{
		registry:     registry,
		registryAddr: registryAddr,
		servers:      make(map[string]*server),
		logger:       logger,
	}
}

//Change 服务器发生变更
func (s *rspServer) Change(u *watcher.ContentChangeArgs) {
	if s.done {
		return
	}
	switch u.OP {
	case watcher.ADD, watcher.CHANGE:
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			//初始化服务器配置
			conf, err := conf.NewServerConf(u.Path, u.Content, u.Version, s.registry)
			if err != nil {
				s.logger.Error(err)
				return
			}

			if _, ok := s.servers[u.Path]; !ok {
				//添加新服务器
				server := newServer(conf, s.registryAddr, s.registry)
				server.logger.Infof("开始启动...")
				if err = server.Start(); err != nil {
					server.logger.Errorf("启动失败 %v", err)
					return
				}
				s.servers[u.Path] = server
				server.logger.Infof("启动成功(%s,%d)", server.GetAddress(), len(server.GetServices()))
			} else {
				//修改服务器
				server := s.servers[u.Path]
				if !conf.IsStop() {
					if err = server.Notify(conf); err != nil {
						server.logger.Errorf("未完成更新 %v", err)
					} else {
						server.logger.Info("配置更新成功")
					}
				}

				if conf.IsStop() || server.GetStatus() != servers.ST_RUNNING {
					server.logger.Info("关闭服务器")
					server.Shutdown()
					delete(s.servers, u.Path)
					return
				}
			}
		}()

	case watcher.DEL:
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if server, ok := s.servers[u.Path]; ok {
				server.logger.Errorf("%s配置已删除", u.Path)
				server.Shutdown()
				server.logger.Info("服务器已关闭")
				delete(s.servers, u.Path)

				return
			}
		}()
	}
}

//Change 服务器发生变更
func (s *rspServer) Shutdown() {
	s.done = true
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, server := range s.servers {
		server.Shutdown()
	}
}
