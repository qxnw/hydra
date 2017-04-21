package web

import (
	"fmt"
)

func (s *WebServer) registryServer() (err error) {
	if s.registry != nil {
		addr := s.GetAddress()
		s.clusterPath, err = s.registry.RegisterWithPath(fmt.Sprintf("%s/%s", s.registryRoot, s.ip), addr)
		return
	}
	return
}
func (s *WebServer) unregistryServer() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)

	}
}
