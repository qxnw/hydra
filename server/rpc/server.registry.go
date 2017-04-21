package rpc

import "fmt"

func (s *RPCServer) registerService() (err error) {
	if s.registry != nil {
		addr := s.GetAddress()
		s.clusterPath, err = s.registry.RegisterWithPath(fmt.Sprintf("%s/%s", s.registryRoot, s.ip), addr)
	}
	return nil
}
func (s *RPCServer) unRegisterService() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
}
